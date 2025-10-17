import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

export default buildModule("LoteTracingModule", (m) => {
  // Contract parameters
  const loteId = m.getParameter("loteId", "LOT-2024-001");
  const temperaturaMinima = m.getParameter("temperaturaMinima", 2);
  const temperaturaMaxima = m.getParameter("temperaturaMaxima", 8);

  // Deploy the LoteTracing PoC contract
  const loteTracing = m.contract("LoteDeProductoTrazablePoC", [
    loteId,
    temperaturaMinima,
    temperaturaMaxima
  ]);

  return { 
    loteTracing,
    contractAddress: loteTracing,
    loteId,
    temperaturaMinima,
    temperaturaMaxima
  };
});