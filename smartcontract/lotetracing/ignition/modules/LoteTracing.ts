import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

export default buildModule("LoteTracingModule", (m) => {
  // Contract parameters
  const sku = m.getParameter("sku", "MED-001");
  const loteId = m.getParameter("loteId", "LOT-2024-001");
  const fechaVencimiento = m.getParameter("fechaVencimiento", BigInt(Math.floor(Date.now() / 1000) + 365 * 24 * 60 * 60)); // 1 year from now
  const temperaturaMinima = m.getParameter("temperaturaMinima", 2);
  const temperaturaMaxima = m.getParameter("temperaturaMaxima", 8);

  // Deploy the LoteTracing contract
  const loteTracing = m.contract("LoteDeProductoTrazable", [
    sku,
    loteId,
    fechaVencimiento,
    temperaturaMinima,
    temperaturaMaxima
  ]);

  // Optional: Authorize a sensor after deployment
  const sensorAddress = m.getParameter("sensorAddress", "0x0000000000000000000000000000000000000000");
  
  // Only authorize sensor if a valid address is provided
  m.call(loteTracing, "gestionarSensor", [sensorAddress, true], {
    id: "authorize_sensor",
    // Only execute if sensor address is not zero address
    after: [loteTracing],
  });

  return { 
    loteTracing,
    contractAddress: loteTracing,
    sku,
    loteId,
    fechaVencimiento,
    temperaturaMinima,
    temperaturaMaxima
  };
});