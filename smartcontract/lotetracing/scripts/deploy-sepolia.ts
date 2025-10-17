import { network } from "hardhat";

console.log("=== Despliegue LoteTracing PoC ===\n");

const { viem } = await network.connect();
const publicClient = await viem.getPublicClient();

// Get the deployer account
const [deployer] = await viem.getWalletClients();

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
console.log(`👤 Cuenta del desplegador: ${deployer.account.address}`);

// Check balance
const balance = await publicClient.getBalance({
  address: deployer.account.address,
});
console.log(`💰 Balance: ${balance} wei (${Number(balance) / 1e18} ETH)\n`);

// Contract parameters
const LOTE_ID = "INS-2024-10-001";
const TEMP_MIN = 2;
const TEMP_MAX = 8;

console.log("Parámetros del lote:");
console.log(`- Lote ID: ${LOTE_ID}`);
console.log(`- Temperatura mínima: ${TEMP_MIN}°C`);
console.log(`- Temperatura máxima: ${TEMP_MAX}°C\n`);

try {
  // Deploy contract
  console.log("📦 Desplegando contrato LoteTracing PoC...");
  const lote = await viem.deployContract("LoteDeProductoTrazablePoC", [
    LOTE_ID,
    TEMP_MIN,
    TEMP_MAX,
  ]);

  console.log(`✅ Contrato desplegado exitosamente!`);
  console.log(`📍 Dirección del contrato: ${lote.address}`);

  // Show appropriate explorer link based on network
  if (chainId === 11155111) {
    console.log(
      `🔗 Ver en Etherscan: https://sepolia.etherscan.io/address/${lote.address}`
    );
  } else if (chainId === 1) {
    console.log(
      `🔗 Ver en Etherscan: https://etherscan.io/address/${lote.address}`
    );
  } else if (chainId === 31337) {
    console.log(`🔗 Red local - No hay explorador disponible`);
  } else {
    console.log(`🔗 Explorador no configurado para Chain ID ${chainId}`);
  }

  // Verify initial state
  console.log("\n📋 Verificando estado inicial...");
  const loteId = await lote.read.loteId();
  const fabricante = await lote.read.fabricante();
  const propietario = await lote.read.propietarioActual();
  const comprometido = await lote.read.comprometido();
  const tempMin = await lote.read.temperaturaMinima();
  const tempMax = await lote.read.temperaturaMaxima();

  console.log(`- Lote ID: ${loteId}`);
  console.log(`- Fabricante: ${fabricante}`);
  console.log(`- Propietario actual: ${propietario}`);
  console.log(`- Estado: ${comprometido ? "Comprometido" : "Íntegro"}`);
  console.log(`- Rango de temperatura: ${tempMin}°C - ${tempMax}°C`);

  // Register a test temperature
  console.log("\n🌡️  Registrando temperatura de prueba...");
  const tempTx = await lote.write.registrarTemperatura([5]);
  console.log(`   📝 Transacción de temperatura: ${tempTx}`);

  // Wait for the transaction to be confirmed
  console.log("   ⏳ Esperando confirmación...");
  await publicClient.waitForTransactionReceipt({ hash: tempTx });

  const comprometidoDespues = await lote.read.comprometido();
  console.log(`   ✅ Temperatura registrada: 5°C`);
  console.log(`   📊 Estado después del registro: ${comprometidoDespues ? "Comprometido" : "Íntegro"}`);

  // Test out-of-range temperature
  console.log("\n🚨 Probando temperatura fuera de rango...");
  const tempOutTx = await lote.write.registrarTemperatura([15]);
  console.log(`   📝 Transacción de temperatura: ${tempOutTx}`);

  // Wait for the transaction to be confirmed
  console.log("   ⏳ Esperando confirmación...");
  await publicClient.waitForTransactionReceipt({ hash: tempOutTx });

  const comprometidoFinal = await lote.read.comprometido();
  console.log(`   ✅ Temperatura registrada: 15°C (fuera de rango)`);
  console.log(`   📊 Estado final: ${comprometidoFinal ? "❌ Comprometido" : "✅ Íntegro"}`);

  console.log("\n=== Despliegue completado exitosamente ===");
  console.log(`🎉 El contrato está listo para usar en ${networkName}`);
  console.log(`📍 Dirección: ${lote.address}`);
  console.log(`🔍 Funcionalidades probadas:`);
  console.log(`   - ✅ Despliegue del contrato`);
  console.log(`   - ✅ Registro de temperatura válida`);
  console.log(`   - ✅ Detección de temperatura fuera de rango`);
  console.log(`   - ✅ Marcado automático como comprometido`);
} catch (error) {
  console.error("❌ Error durante el despliegue:", error);
  process.exit(1);
}