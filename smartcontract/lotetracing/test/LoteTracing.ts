import assert from "node:assert/strict";
import { describe, it } from "node:test";

import { network } from "hardhat";

describe("LoteTracing PoC", async function () {
  const { viem } = await network.connect();
  const publicClient = await viem.getPublicClient();

  // Test addresses
  const [fabricante, distribuidor, farmacia] = await viem.getWalletClients();

  // Contract parameters
  const LOTE_ID = "LOT-2024-001";
  const TEMP_MIN = 2;
  const TEMP_MAX = 8;

  it("Should deploy and initialize correctly", async function () {
    const lote = await viem.deployContract("LoteTracing", [
      LOTE_ID,
      TEMP_MIN,
      TEMP_MAX,
    ]);

    const loteId = await lote.read.loteId();
    const fabricanteAddr = await lote.read.fabricante();
    const propietarioActual = await lote.read.propietarioActual();
    const comprometido = await lote.read.comprometido();

    assert.equal(loteId, LOTE_ID);
    assert.equal(fabricanteAddr.toLowerCase(), fabricante.account.address.toLowerCase());
    assert.equal(propietarioActual.toLowerCase(), fabricante.account.address.toLowerCase());
    assert.equal(comprometido, false);
  });

  it("Should register valid temperature correctly", async function () {
    const lote = await viem.deployContract("LoteTracing", [
      LOTE_ID,
      TEMP_MIN,
      TEMP_MAX,
    ]);

    // Register valid temperature
    const temperatura = 5;
    await lote.write.registrarTemperatura([temperatura, TEMP_MIN, TEMP_MAX]);

    const comprometido = await lote.read.comprometido();
    assert.equal(comprometido, false);
  });

  it("Should mark lot as compromised when temperature is out of range", async function () {
    const lote = await viem.deployContract("LoteTracing", [
      LOTE_ID,
      TEMP_MIN,
      TEMP_MAX,
    ]);

    // Register temperature out of range
    const temperaturaAlta = 15; // Above TEMP_MAX (8)
    await lote.write.registrarTemperatura([temperaturaAlta, TEMP_MIN, TEMP_MAX]);

    const comprometido = await lote.read.comprometido();
    assert.equal(comprometido, true);
  });

  it("Should transfer custody correctly", async function () {
    const lote = await viem.deployContract("LoteTracing", [
      LOTE_ID,
      TEMP_MIN,
      TEMP_MAX,
    ]);

    // Transfer to distributor
    await lote.write.transferirCustodia([distribuidor.account.address]);

    const nuevoPropietario = await lote.read.propietarioActual();
    assert.equal(nuevoPropietario.toLowerCase(), distribuidor.account.address.toLowerCase());
  });

  it("Should complete full traceability cycle", async function () {
    const lote = await viem.deployContract("LoteTracing", [
      LOTE_ID,
      TEMP_MIN,
      TEMP_MAX,
    ]);

    // 1. Register valid temperatures as fabricante
    await lote.write.registrarTemperatura([4, TEMP_MIN, TEMP_MAX]);
    await lote.write.registrarTemperatura([6, TEMP_MIN, TEMP_MAX]);

    // 2. Transfer custody: Fabricante -> Distribuidor
    await lote.write.transferirCustodia([distribuidor.account.address]);

    // 3. Register temperature as distribuidor
    await distribuidor.writeContract({
      address: lote.address,
      abi: lote.abi,
      functionName: "registrarTemperatura",
      args: [5, TEMP_MIN, TEMP_MAX]
    });

    // 4. Transfer custody: Distribuidor -> Farmacia
    await distribuidor.writeContract({
      address: lote.address,
      abi: lote.abi,
      functionName: "transferirCustodia",
      args: [farmacia.account.address]
    });

    // Verify final state
    const propietarioFinal = await lote.read.propietarioActual();
    const comprometido = await lote.read.comprometido();

    assert.equal(propietarioFinal.toLowerCase(), farmacia.account.address.toLowerCase());
    assert.equal(comprometido, false);
  });

  it("Should reject unauthorized operations", async function () {
    const lote = await viem.deployContract("LoteTracing", [
      LOTE_ID,
      TEMP_MIN,
      TEMP_MAX,
    ]);

    // Try to register temperature from non-owner
    await assert.rejects(
      distribuidor.writeContract({
        address: lote.address,
        abi: lote.abi,
        functionName: "registrarTemperatura",
        args: [5, TEMP_MIN, TEMP_MAX]
      }),
      /Accion solo permitida para el propietario actual/
    );

    // Try to transfer custody from non-owner
    await assert.rejects(
      distribuidor.writeContract({
        address: lote.address,
        abi: lote.abi,
        functionName: "transferirCustodia",
        args: [farmacia.account.address]
      }),
      /Accion solo permitida para el propietario actual/
    );
  });

  it("Should prevent temperature registration on compromised lot", async function () {
    const lote = await viem.deployContract("LoteTracing", [
      LOTE_ID,
      TEMP_MIN,
      TEMP_MAX,
    ]);

    // Compromise the lot
    await lote.write.registrarTemperatura([15, TEMP_MIN, TEMP_MAX]); // Out of range

    const comprometido = await lote.read.comprometido();
    assert.equal(comprometido, true);

    // Try to register another temperature
    await assert.rejects(
      lote.write.registrarTemperatura([5, TEMP_MIN, TEMP_MAX]),
      /El lote ya esta comprometido/
    );
  });

  it("Should emit events correctly", async function () {
    const deploymentBlockNumber = await publicClient.getBlockNumber();
    
    const lote = await viem.deployContract("LoteTracing", [
      LOTE_ID,
      TEMP_MIN,
      TEMP_MAX,
    ]);

    // Transfer custody to trigger event
    await lote.write.transferirCustodia([distribuidor.account.address]);

    // Register out-of-range temperature to trigger compromised event
    await distribuidor.writeContract({
      address: lote.address,
      abi: lote.abi,
      functionName: "registrarTemperatura",
      args: [15, TEMP_MIN, TEMP_MAX]
    });

    // Check for custody transfer event
    const custodyEvents = await publicClient.getContractEvents({
      address: lote.address,
      abi: lote.abi,
      eventName: "CustodiaTransferida",
      fromBlock: deploymentBlockNumber,
      strict: true,
    });

    assert.equal(custodyEvents.length, 1);

    // Check for compromised event
    const compromisedEvents = await publicClient.getContractEvents({
      address: lote.address,
      abi: lote.abi,
      eventName: "LoteComprometido",
      fromBlock: deploymentBlockNumber,
      strict: true,
    });

    assert.equal(compromisedEvents.length, 1);
  });
});