import { network } from "hardhat";

console.log("=== Demo Función crearNuevoLote() ===\n");

const networkConnection = await network.connect();
const { viem } = networkConnection as any;
const publicClient = await viem.getPublicClient();

// Get wallet clients
const walletClients = await viem.getWalletClients();
const [fabricante, usuario1, usuario2] = walletClients;

console.log("Actores del demo:");
console.log(`- Fabricante: ${fabricante.account.address}`);
console.log(`- Usuario 1: ${usuario1.account.address}`);
console.log(`- Usuario 2: ${usuario2.account.address}\n`);

// Initial contract parameters
const LOTE_INICIAL = "LOT-INICIAL-001";
const TEMP_MIN_INICIAL = 2;
const TEMP_MAX_INICIAL = 8;

console.log("1. 📦 Desplegando contrato inicial...");
const lote = await viem.deployContract("LoteTracing", [
  LOTE_INICIAL,
  TEMP_MIN_INICIAL,
  TEMP_MAX_INICIAL,
]);

console.log(`   ✅ Contrato desplegado en: ${lote.address}`);
console.log(`   📋 Lote inicial: ${LOTE_INICIAL}`);
console.log(`   🌡️  Rango inicial: ${TEMP_MIN_INICIAL}°C - ${TEMP_MAX_INICIAL}°C\n`);

// Register some temperatures and transfer custody
console.log("2. 🔄 Operaciones iniciales...");
await lote.write.registrarTemperatura([3, 7]);
await lote.write.transferirCustodia([usuario1.account.address]);
console.log(`   ✅ Temperatura registrada y custodia transferida a Usuario 1\n`);

// Demonstrate crearNuevoLote by different users
const lotes = [
  {
    usuario: usuario1,
    nombre: "Usuario 1",
    loteId: "LOT-USER1-001",
    tempMin: -10,
    tempMax: 25,
  },
  {
    usuario: usuario2,
    nombre: "Usuario 2", 
    loteId: "LOT-USER2-001",
    tempMin: 0,
    tempMax: 10,
  },
  {
    usuario: fabricante,
    nombre: "Fabricante",
    loteId: "LOT-FABRICANTE-002",
    tempMin: 5,
    tempMax: 15,
  }
];

for (let i = 0; i < lotes.length; i++) {
  const { usuario, nombre, loteId, tempMin, tempMax } = lotes[i];
  
  console.log(`${i + 3}. 🆕 ${nombre} crea nuevo lote...`);
  console.log(`   📦 Nuevo lote: ${loteId}`);
  console.log(`   🌡️  Nuevo rango: ${tempMin}°C - ${tempMax}°C`);
  
  const hash = await usuario.writeContract({
    address: lote.address,
    abi: lote.abi,
    functionName: "crearNuevoLote",
    args: [loteId, tempMin, tempMax],
  });
  await publicClient.waitForTransactionReceipt({ hash });
  
  // Verify state after creation
  const estadoActual = {
    loteId: await lote.read.loteId(),
    tempMin: await lote.read.temperaturaMinima(),
    tempMax: await lote.read.temperaturaMaxima(),
    propietario: await lote.read.propietarioActual(),
    comprometido: await lote.read.comprometido(),
    tempRegMin: await lote.read.tempRegMinima(),
    tempRegMax: await lote.read.tempRegMaxima(),
  };
  
  console.log(`   ✅ Estado después de la creación:`);
  console.log(`      - Lote ID: ${estadoActual.loteId}`);
  console.log(`      - Rango permitido: ${estadoActual.tempMin}°C - ${estadoActual.tempMax}°C`);
  console.log(`      - Propietario: ${estadoActual.propietario}`);
  console.log(`      - Comprometido: ${estadoActual.comprometido}`);
  console.log(`      - Temp. registradas: ${estadoActual.tempRegMin}°C - ${estadoActual.tempRegMax}°C`);
  
  // Test temperature registration with new parameters
  const testTemp = [tempMin + 1, tempMax - 1];
  await usuario.writeContract({
    address: lote.address,
    abi: lote.abi,
    functionName: "registrarTemperatura",
    args: testTemp,
  });
  
  const comprometidoDespues = await lote.read.comprometido();
  console.log(`   🌡️  Prueba de temperatura [${testTemp[0]}, ${testTemp[1]}]: ${comprometidoDespues ? "Comprometido" : "Válido"}\n`);
}

console.log("=== Resumen de Funcionalidad crearNuevoLote() ===");
console.log("✅ Características demostradas:");
console.log("   - Cualquier usuario puede crear un nuevo lote");
console.log("   - Se reinician todos los parámetros del lote");
console.log("   - El creador se convierte en el nuevo propietario");
console.log("   - Se pueden establecer nuevos rangos de temperatura");
console.log("   - El estado se resetea completamente (no comprometido)");
console.log("   - Las temperaturas registradas se reinician a 0");
console.log("\n⚠️  Consideraciones de seguridad:");
console.log("   - Esta función rompe la inmutabilidad del contrato");
console.log("   - Permite sobrescribir el historial del lote");
console.log("   - No hay restricciones de acceso");
console.log("   - Podría comprometer la trazabilidad");

console.log(`\n📍 Dirección del contrato: ${lote.address}`);
console.log("🎉 Demo completado exitosamente");