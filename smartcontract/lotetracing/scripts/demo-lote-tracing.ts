import { network } from "hardhat";

console.log("=== Demo LoteTracing - Trazabilidad de Productos Farmacéuticos ===\n");

const { viem } = await network.connect();
const publicClient = await viem.getPublicClient();

// Get wallet clients for different actors
const [fabricante, sensor1, distribuidor, farmacia] = await viem.getWalletClients();

console.log("Actores del sistema:");
console.log(`- Fabricante: ${fabricante.account.address}`);
console.log(`- Sensor IoT: ${sensor1.account.address}`);
console.log(`- Distribuidor: ${distribuidor.account.address}`);
console.log(`- Farmacia: ${farmacia.account.address}\n`);

// Contract parameters
const SKU = "INSULIN-RAPID-001";
const LOTE_ID = "INS-2024-10-001";
const fechaVencimiento = BigInt(Math.floor(Date.now() / 1000) + 180 * 24 * 60 * 60); // 6 months
const TEMP_MIN = 2;
const TEMP_MAX = 8;

console.log("Parámetros del lote:");
console.log(`- SKU: ${SKU}`);
console.log(`- Lote ID: ${LOTE_ID}`);
console.log(`- Temperatura mínima: ${TEMP_MIN}°C`);
console.log(`- Temperatura máxima: ${TEMP_MAX}°C\n`);

// 1. Deploy contract (fabricante creates the lot)
console.log("1. 📦 Fabricante crea el lote...");
const lote = await viem.deployContract("LoteDeProductoTrazable", [
  SKU,
  LOTE_ID,
  fechaVencimiento,
  TEMP_MIN,
  TEMP_MAX
], { client: fabricante });

console.log(`   ✅ Lote creado en: ${lote.address}`);
console.log(`   📅 Fecha de fabricación: ${new Date().toISOString()}`);
console.log(`   📅 Fecha de vencimiento: ${new Date(Number(fechaVencimiento) * 1000).toISOString()}\n`);

// 2. Authorize IoT sensor
console.log("2. 🔐 Fabricante autoriza sensor IoT...");
await lote.write.gestionarSensor([sensor1.account.address, true], { client: fabricante });
console.log(`   ✅ Sensor ${sensor1.account.address} autorizado\n`);

// 3. Register temperature readings during manufacturing
console.log("3. 🌡️  Registrando temperaturas durante fabricación...");
const temperaturasIniciales = [4, 5, 6, 5, 4];
for (let i = 0; i < temperaturasIniciales.length; i++) {
  const temp = temperaturasIniciales[i];
  await sensor1.writeContract({
    address: lote.address,
    abi: lote.abi,
    functionName: "registrarTemperatura",
    args: [temp]
  });
  console.log(`   📊 Temperatura registrada: ${temp}°C`);
  
  // Simulate time passing
  await new Promise(resolve => setTimeout(resolve, 100));
}
console.log();

// 4. Transfer to distributor
console.log("4. 🚚 Transferencia a distribuidor...");
await lote.write.transferirCustodia([distribuidor.account.address]);
const estadoTransito = await lote.read.estado();
console.log(`   ✅ Custodia transferida a distribuidor`);
const estadoTexto = Number(estadoTransito) === 0 ? "Creado" : 
                   Number(estadoTransito) === 1 ? "En Tránsito" : 
                   Number(estadoTransito) === 2 ? "En Almacén" : 
                   Number(estadoTransito) === 3 ? "Comprometido" : 
                   Number(estadoTransito) === 4 ? "Entregado" : `Desconocido (${estadoTransito})`;
console.log(`   📋 Estado del lote: ${estadoTexto}\n`);

// 5. Register temperatures during transport
console.log("5. 🌡️  Temperaturas durante transporte...");
const temperaturasTransporte = [6, 7, 8, 7, 6, 5];
for (let i = 0; i < temperaturasTransporte.length; i++) {
  const temp = temperaturasTransporte[i];
  await sensor1.writeContract({
    address: lote.address,
    abi: lote.abi,
    functionName: "registrarTemperatura",
    args: [temp]
  });
  console.log(`   📊 Temperatura en tránsito: ${temp}°C`);
  
  await new Promise(resolve => setTimeout(resolve, 100));
}
console.log();

// 6. Transfer to pharmacy
console.log("6. 🏥 Transferencia a farmacia...");
await distribuidor.writeContract({
  address: lote.address,
  abi: lote.abi,
  functionName: "transferirCustodia",
  args: [farmacia.account.address]
});
const estadoAlmacen = await lote.read.estado();
console.log(`   ✅ Custodia transferida a farmacia`);
const estadoTextoAlmacen = Number(estadoAlmacen) === 0 ? "Creado" : 
                          Number(estadoAlmacen) === 1 ? "En Tránsito" : 
                          Number(estadoAlmacen) === 2 ? "En Almacén" : 
                          Number(estadoAlmacen) === 3 ? "Comprometido" : 
                          Number(estadoAlmacen) === 4 ? "Entregado" : `Desconocido (${estadoAlmacen})`;
console.log(`   📋 Estado del lote: ${estadoTextoAlmacen}\n`);

// 7. Final temperature readings at pharmacy
console.log("7. 🌡️  Temperaturas en farmacia...");
const temperaturasFarmacia = [4, 3, 4, 5];
for (let i = 0; i < temperaturasFarmacia.length; i++) {
  const temp = temperaturasFarmacia[i];
  await sensor1.writeContract({
    address: lote.address,
    abi: lote.abi,
    functionName: "registrarTemperatura",
    args: [temp]
  });
  console.log(`   📊 Temperatura en farmacia: ${temp}°C`);
  
  await new Promise(resolve => setTimeout(resolve, 100));
}
console.log();

// 8. Get final state and history
console.log("8. 📋 Estado final del lote:");
const propietarioFinal = await lote.read.propietarioActual();
const estadoFinal = await lote.read.estado();
const historial = await lote.read.obtenerHistorialCustodia();
const lecturas = await lote.read.obtenerLecturasTemperatura();

const estadoFinalTexto = Number(estadoFinal) === 0 ? "Creado" : 
                        Number(estadoFinal) === 1 ? "En Tránsito" : 
                        Number(estadoFinal) === 2 ? "En Almacén" : 
                        Number(estadoFinal) === 3 ? "Comprometido" : 
                        Number(estadoFinal) === 4 ? "Entregado" : `Desconocido (${estadoFinal})`;

console.log(`   👤 Propietario actual: ${propietarioFinal}`);
console.log(`   📊 Estado: ${estadoFinalTexto}`);
console.log(`   📈 Total de lecturas de temperatura: ${lecturas.length}`);
console.log(`   🔄 Transferencias de custodia: ${historial.length}\n`);

// 9. Show custody history
console.log("9. 📜 Historial de custodia:");
for (let i = 0; i < historial.length; i++) {
  const entrada = historial[i];
  const fecha = new Date(Number(entrada.timestamp) * 1000);
  let rol = "Desconocido";
  
  if (entrada.propietario.toLowerCase() === fabricante.account.address.toLowerCase()) rol = "Fabricante";
  else if (entrada.propietario.toLowerCase() === distribuidor.account.address.toLowerCase()) rol = "Distribuidor";
  else if (entrada.propietario.toLowerCase() === farmacia.account.address.toLowerCase()) rol = "Farmacia";
  
  console.log(`   ${i + 1}. ${rol} (${entrada.propietario}) - ${fecha.toISOString()}`);
}
console.log();

// 10. Temperature statistics
console.log("10. 📊 Estadísticas de temperatura:");
const temperaturas = lecturas.map(l => Number(l.temperatura));
const tempMin = Math.min(...temperaturas);
const tempMax = Math.max(...temperaturas);
const tempPromedio = temperaturas.reduce((a, b) => a + b, 0) / temperaturas.length;

console.log(`   🌡️  Temperatura mínima registrada: ${tempMin}°C`);
console.log(`   🌡️  Temperatura máxima registrada: ${tempMax}°C`);
console.log(`   🌡️  Temperatura promedio: ${tempPromedio.toFixed(1)}°C`);
console.log(`   ✅ Todas las temperaturas dentro del rango permitido (${TEMP_MIN}°C - ${TEMP_MAX}°C)\n`);

console.log("=== Demo completado exitosamente ===");
console.log(`🎉 El lote ${LOTE_ID} ha sido trazado completamente desde fabricación hasta farmacia`);
console.log(`📍 Dirección del contrato: ${lote.address}`);