import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

export default buildModule("LoteTracingModule", (m) => {
  // Contract parameters with updated defaults
  const loteId = m.getParameter("loteId", "LOT-2024-001");
  const temperaturaMinima = m.getParameter("temperaturaMinima", 2);
  const temperaturaMaxima = m.getParameter("temperaturaMaxima", 8);

  // Deploy the LoteTracing PoC contract
  // Note: Contract now includes crearNuevoLote() function for lot reinitialization
  const loteTracing = m.contract("LoteTracing", [
    loteId,
    temperaturaMinima,
    temperaturaMaxima,
  ]);

  return {
    loteTracing,
  };
});
