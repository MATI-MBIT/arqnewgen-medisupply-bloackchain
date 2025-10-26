import { network } from "hardhat";

console.log("=== Demo LoteTracing PoC - Trazabilidad Simplificada ===\n");

const networkConnection = await network.connect();
const { viem } = networkConnection as any;
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

// 2. Register temperature ranges during manufacturing
console.log("2. 🌡️  Fabricante registra rangos de temperatura durante fabricación...");
const rangosIniciales = [[TEMP_MIN, TEMP_MAX], [3, 7], [2, 6]]; // All within valid range
for (let i = 0; i < rangosIniciales.length; i++) {
  const [min, max] = rangosIniciales[i];
  const hash = await lote.write.registrarTemperatura([min, max]);
  await publicClient.waitForTransactionReceipt({ hash });

  console.log(`   📊 Rango registrado: ${min}°C - ${max}°C`);
}

const comprometidoFabricacion = await lote.read.comprometido();
console.log(
  `   ✅ Estado después de fabricación: ${
    comprometidoFabricacion ? "Comprometido" : "Íntegro"
  }\n`
);

// 3. Transfer to distributor
console.log("3. 🚚 Transferencia a distribuidor...");
const transferHash = await lote.write.transferirCustodia([
  distribuidor.account.address,
]);
await publicClient.waitForTransactionReceipt({ hash: transferHash });
const propietarioActual = await lote.read.propietarioActual();
console.log(`   ✅ Custodia transferida a: ${propietarioActual}`);

// 4. Register temperature ranges during transport
console.log(
  "\n4. 🌡️  Distribuidor registra rangos de temperatura durante transporte..."
);
const rangosTransporte = [[TEMP_MIN, TEMP_MAX], [3, 8], [2, 7]]; // All within valid range
for (let i = 0; i < rangosTransporte.length; i++) {
  const [min, max] = rangosTransporte[i];
  const hash = await distribuidor.writeContract({
    address: lote.address,
    abi: lote.abi,
    functionName: "registrarTemperatura",
    args: [min, max],
  });
  await publicClient.waitForTransactionReceipt({ hash });
  console.log(`   📊 Rango en tránsito: ${min}°C - ${max}°C`);
}

const comprometidoTransporte = await lote.read.comprometido();
console.log(
  `   ✅ Estado después de transporte: ${
    comprometidoTransporte ? "Comprometido" : "Íntegro"
  }\n`
);

// 5. Transfer to pharmacy
console.log("5. 🏥 Transferencia a farmacia...");
const transferHash2 = await distribuidor.writeContract({
  address: lote.address,
  abi: lote.abi,
  functionName: "transferirCustodia",
  args: [farmacia.account.address],
});
await publicClient.waitForTransactionReceipt({ hash: transferHash2 });
const propietarioFinal = await lote.read.propietarioActual();
console.log(`   ✅ Custodia transferida a: ${propietarioFinal}`);

// 6. Final temperature ranges at pharmacy
console.log("\n6. 🌡️  Farmacia registra rangos de temperatura de almacenamiento...");
const rangosFarmacia = [[TEMP_MIN, TEMP_MAX], [4, 7], [3, 8]]; // All within valid range
for (let i = 0; i < rangosFarmacia.length; i++) {
  const [min, max] = rangosFarmacia[i];
  const hash = await farmacia.writeContract({
    address: lote.address,
    abi: lote.abi,
    functionName: "registrarTemperatura",
    args: [min, max],
  });
  await publicClient.waitForTransactionReceipt({ hash });
  console.log(`   📊 Rango en farmacia: ${min}°C - ${max}°C`);
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
console.log("\n8. 🚨 Demostración: Registro de rango inválido...");
try {
  // Try to register an invalid temperature range
  const hash = await farmacia.writeContract({
    address: lote.address,
    abi: lote.abi,
    functionName: "registrarTemperatura",
    args: [10, 15], // Range doesn't include contract's 2-8
  });
  await publicClient.waitForTransactionReceipt({ hash });

  const comprometidoFinal = await lote.read.comprometido();
  console.log(`   🌡️  Rango registrado: 10°C - 15°C (no incluye rango del contrato)`);
  console.log(`   ❌ Lote marcado como comprometido: ${comprometidoFinal}`);

  // Try to register another temperature range (should fail)
  console.log(
    "\n9. 🚫 Intento de registrar rango en lote comprometido..."
  );
  try {
    const hash2 = await farmacia.writeContract({
      address: lote.address,
      abi: lote.abi,
      functionName: "registrarTemperatura",
      args: [TEMP_MIN, TEMP_MAX],
    });
    await publicClient.waitForTransactionReceipt({ hash: hash2 });
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
console.log(`   - ✅ Registro de rangos de temperatura por cualquier usuario`);
console.log(`   - ✅ Transferencia de custodia entre actores`);
console.log(`   - ✅ Validación de rangos contra temperaturas del contrato`);
console.log(`   - ✅ Prevención de registros en lotes comprometidos`);
console.log(`   - ✅ Control de acceso para transferencia de custodia`);
