import { network } from "hardhat";

console.log("=== Demo LoteTracing PoC - Trazabilidad Simplificada ===\n");

const { viem } = await network.connect();
const publicClient = await viem.getPublicClient();

// Get wallet clients - handle both local and testnet networks
const walletClients = await viem.getWalletClients();
const chainId = await publicClient.getChainId();
const networkName =
  chainId === 31337
    ? "hardhat"
    : chainId === 11155111
    ? "sepolia"
    : chainId === 1
    ? "mainnet"
    : `unknown-${chainId}`;

console.log(`🌐 Red: ${networkName} (Chain ID: ${chainId})`);

let fabricante, distribuidor, farmacia;

if (
  networkName === "hardhat" ||
  networkName === "localhost" ||
  walletClients.length >= 3
) {
  // Local network - multiple wallets available
  [fabricante, distribuidor, farmacia] = walletClients;
} else {
  // Testnet - use the same wallet for all actors (for demo purposes)
  fabricante = walletClients[0];
  distribuidor = walletClients[0];
  farmacia = walletClients[0];

  console.log(
    "⚠️  Nota: En testnet se usa la misma cuenta para todos los actores (solo para demo)"
  );
}

console.log("Actores del sistema:");
console.log(`- Fabricante: ${fabricante.account.address}`);
console.log(`- Distribuidor: ${distribuidor.account.address}`);
console.log(`- Farmacia: ${farmacia.account.address}\n`);

// Contract parameters
const LOTE_ID = "INS-2024-10-001";
const TEMP_MIN = 2;
const TEMP_MAX = 8;

console.log("Parámetros del lote:");
console.log(`- Lote ID: ${LOTE_ID}`);
console.log(`- Temperatura mínima: ${TEMP_MIN}°C`);
console.log(`- Temperatura máxima: ${TEMP_MAX}°C\n`);

// 1. Deploy contract (fabricante creates the lot)
console.log("1. 📦 Fabricante crea el lote...");
const lote = await viem.deployContract("LoteTracing", [
  LOTE_ID,
  TEMP_MIN,
  TEMP_MAX,
]);

console.log(`   ✅ Lote creado en: ${lote.address}`);
console.log(`   📅 Fecha de creación: ${new Date().toISOString()}\n`);

// 2. Register temperature readings during manufacturing
console.log("2. 🌡️  Fabricante registra temperaturas durante fabricación...");
const temperaturasIniciales = [4, 5, 6, 5, 4];
for (let i = 0; i < temperaturasIniciales.length; i++) {
  const temp = temperaturasIniciales[i];
  await lote.write.registrarTemperatura([temp, TEMP_MIN, TEMP_MAX]);
  console.log(`   📊 Temperatura registrada: ${temp}°C`);

  // Simulate time passing
  await new Promise((resolve) => setTimeout(resolve, 100));
}

const comprometidoFabricacion = await lote.read.comprometido();
console.log(
  `   ✅ Estado después de fabricación: ${
    comprometidoFabricacion ? "Comprometido" : "Íntegro"
  }\n`
);

// 3. Transfer to distributor
console.log("3. 🚚 Transferencia a distribuidor...");
await lote.write.transferirCustodia([distribuidor.account.address]);
const propietarioActual = await lote.read.propietarioActual();
console.log(`   ✅ Custodia transferida a: ${propietarioActual}`);

// 4. Register temperatures during transport
console.log(
  "\n4. 🌡️  Distribuidor registra temperaturas durante transporte..."
);
const temperaturasTransporte = [6, 7, 8, 7, 6];
for (let i = 0; i < temperaturasTransporte.length; i++) {
  const temp = temperaturasTransporte[i];
  await distribuidor.writeContract({
    address: lote.address,
    abi: lote.abi,
    functionName: "registrarTemperatura",
    args: [temp, TEMP_MIN, TEMP_MAX],
  });
  console.log(`   📊 Temperatura en tránsito: ${temp}°C`);

  await new Promise((resolve) => setTimeout(resolve, 100));
}

const comprometidoTransporte = await lote.read.comprometido();
console.log(
  `   ✅ Estado después de transporte: ${
    comprometidoTransporte ? "Comprometido" : "Íntegro"
  }\n`
);

// 5. Transfer to pharmacy
console.log("5. 🏥 Transferencia a farmacia...");
await distribuidor.writeContract({
  address: lote.address,
  abi: lote.abi,
  functionName: "transferirCustodia",
  args: [farmacia.account.address],
});
const propietarioFinal = await lote.read.propietarioActual();
console.log(`   ✅ Custodia transferida a: ${propietarioFinal}`);

// 6. Final temperature readings at pharmacy
console.log("\n6. 🌡️  Farmacia registra temperaturas de almacenamiento...");
const temperaturasFarmacia = [4, 3, 4, 5];
for (let i = 0; i < temperaturasFarmacia.length; i++) {
  const temp = temperaturasFarmacia[i];
  await farmacia.writeContract({
    address: lote.address,
    abi: lote.abi,
    functionName: "registrarTemperatura",
    args: [temp, TEMP_MIN, TEMP_MAX],
  });
  console.log(`   📊 Temperatura en farmacia: ${temp}°C`);

  await new Promise((resolve) => setTimeout(resolve, 100));
}

// 7. Get final state
console.log("\n7. 📋 Estado final del lote:");
const estadoFinal = await lote.read.comprometido();
const fabricanteAddr = await lote.read.fabricante();

console.log(`   👤 Fabricante original: ${fabricanteAddr}`);
console.log(`   👤 Propietario actual: ${propietarioFinal}`);
console.log(
  `   📊 Estado final: ${estadoFinal ? "❌ Comprometido" : "✅ Íntegro"}`
);
console.log(`   🌡️  Rango permitido: ${TEMP_MIN}°C - ${TEMP_MAX}°C`);

// 8. Demonstrate compromised scenario
console.log("\n8. 🚨 Demostración: Registro de temperatura fuera de rango...");
try {
  // Try to register an out-of-range temperature
  await farmacia.writeContract({
    address: lote.address,
    abi: lote.abi,
    functionName: "registrarTemperatura",
    args: [15, TEMP_MIN, TEMP_MAX], // Way above TEMP_MAX
  });

  const comprometidoFinal = await lote.read.comprometido();
  console.log(`   🌡️  Temperatura registrada: 15°C (fuera de rango)`);
  console.log(`   ❌ Lote marcado como comprometido: ${comprometidoFinal}`);

  // Try to register another temperature (should fail)
  console.log(
    "\n9. 🚫 Intento de registrar temperatura en lote comprometido..."
  );
  try {
    await farmacia.writeContract({
      address: lote.address,
      abi: lote.abi,
      functionName: "registrarTemperatura",
      args: [5, TEMP_MIN, TEMP_MAX],
    });
  } catch (error) {
    console.log(`   ✅ Registro rechazado correctamente: Lote ya comprometido`);
  }
} catch (error) {
  console.log(
    `   ⚠️  Error en demostración: ${
      error instanceof Error ? error.message : String(error)
    }`
  );
}

console.log("\n=== Demo completado exitosamente ===");
console.log(
  `🎉 El lote ${LOTE_ID} ha sido trazado desde fabricación hasta farmacia`
);
console.log(`📍 Dirección del contrato: ${lote.address}`);
console.log(`🔍 Funcionalidades demostradas:`);
console.log(`   - ✅ Creación de lote con parámetros de temperatura`);
console.log(`   - ✅ Registro de temperaturas por propietario actual`);
console.log(`   - ✅ Transferencia de custodia entre actores`);
console.log(`   - ✅ Detección automática de temperaturas fuera de rango`);
console.log(`   - ✅ Prevención de registros en lotes comprometidos`);
console.log(`   - ✅ Control de acceso basado en propietario actual`);
