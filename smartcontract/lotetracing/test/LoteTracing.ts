import assert from "node:assert/strict";
import { describe, it } from "node:test";

import { network } from "hardhat";

describe("LoteTracing PoC", async function () {
  const networkConnection = await network.connect();
  const { viem } = networkConnection as any;
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
    assert.equal(
      fabricanteAddr.toLowerCase(),
      fabricante.account.address.toLowerCase()
    );
    assert.equal(
      propietarioActual.toLowerCase(),
      fabricante.account.address.toLowerCase()
    );
    assert.equal(comprometido, false);
  });

  it("Should register valid temperature range correctly", async function () {
    const lote = await viem.deployContract("LoteTracing", [
      LOTE_ID,
      TEMP_MIN,
      TEMP_MAX,
    ]);

    // Register valid temperature range within contract's limits
    await lote.write.registrarTemperatura([TEMP_MIN, TEMP_MAX]); // Range 2-8 is within limits

    const comprometido = await lote.read.comprometido();
    assert.equal(comprometido, false);
  });

  it("Should mark lot as compromised when temperature range is invalid", async function () {
    const lote = await viem.deployContract("LoteTracing", [
      LOTE_ID,
      TEMP_MIN,
      TEMP_MAX,
    ]);

    // Register temperature range outside contract's limits (tempMax > 8)
    await lote.write.registrarTemperatura([10, 15]); // Range 10-15 exceeds contract's max of 8

    const comprometido = await lote.read.comprometido();
    assert.equal(comprometido, true);
  });

  it("Should mark lot as compromised when tempMin is below limit", async function () {
    const lote = await viem.deployContract("LoteTracing", [
      LOTE_ID,
      TEMP_MIN,
      TEMP_MAX,
    ]);

    // Register temperature range with tempMin below contract's limit (tempMin < 2)
    await lote.write.registrarTemperatura([0, 6]); // Range 0-6, tempMin=0 < 2

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
    assert.equal(
      nuevoPropietario.toLowerCase(),
      distribuidor.account.address.toLowerCase()
    );
  });

  it("Should complete full traceability cycle", async function () {
    const lote = await viem.deployContract("LoteTracing", [
      LOTE_ID,
      TEMP_MIN,
      TEMP_MAX,
    ]);

    // 1. Register valid temperature ranges as fabricante
    await lote.write.registrarTemperatura([TEMP_MIN, TEMP_MAX]);
    await lote.write.registrarTemperatura([3, 7]); // Another valid range within limits

    // 2. Transfer custody: Fabricante -> Distribuidor
    await lote.write.transferirCustodia([distribuidor.account.address]);

    // 3. Register temperature range as distribuidor
    await distribuidor.writeContract({
      address: lote.address,
      abi: lote.abi,
      functionName: "registrarTemperatura",
      args: [TEMP_MIN, TEMP_MAX],
    });

    // 4. Transfer custody: Distribuidor -> Farmacia
    await distribuidor.writeContract({
      address: lote.address,
      abi: lote.abi,
      functionName: "transferirCustodia",
      args: [farmacia.account.address],
    });

    // Verify final state
    const propietarioFinal = await lote.read.propietarioActual();
    const comprometido = await lote.read.comprometido();

    assert.equal(
      propietarioFinal.toLowerCase(),
      farmacia.account.address.toLowerCase()
    );
    assert.equal(comprometido, false);
  });

  it("Should allow anyone to register temperature but restrict custody transfer", async function () {
    const lote = await viem.deployContract("LoteTracing", [
      LOTE_ID,
      TEMP_MIN,
      TEMP_MAX,
    ]);

    // Anyone can register temperature (this is the intended behavior)
    await distribuidor.writeContract({
      address: lote.address,
      abi: lote.abi,
      functionName: "registrarTemperatura",
      args: [TEMP_MIN, TEMP_MAX],
    });

    // But only the owner can transfer custody
    await assert.rejects(
      distribuidor.writeContract({
        address: lote.address,
        abi: lote.abi,
        functionName: "transferirCustodia",
        args: [farmacia.account.address],
      }),
      /Accion solo permitida para el propietario actual/
    );
  });

  it("Should allow temperature registration even on compromised lot", async function () {
    const lote = await viem.deployContract("LoteTracing", [
      LOTE_ID,
      TEMP_MIN,
      TEMP_MAX,
    ]);

    // Compromise the lot with invalid range
    await lote.write.registrarTemperatura([10, 15]); // Range 10-15 exceeds contract's max of 8

    const comprometido = await lote.read.comprometido();
    assert.equal(comprometido, true);

    // Should still allow temperature registration (validation is commented out)
    await lote.write.registrarTemperatura([TEMP_MIN, TEMP_MAX]);

    // Verify the temperature was registered
    const tempRegMinima = await lote.read.tempRegMinima();
    const tempRegMaxima = await lote.read.tempRegMaxima();
    assert.equal(tempRegMinima, TEMP_MIN);
    assert.equal(tempRegMaxima, TEMP_MAX);
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

    // Register invalid temperature range to trigger compromised event
    await distribuidor.writeContract({
      address: lote.address,
      abi: lote.abi,
      functionName: "registrarTemperatura",
      args: [10, 15], // Range 10-15 exceeds contract's max of 8
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

  it("Should handle edge cases correctly", async function () {
    const lote = await viem.deployContract("LoteTracing", [
      LOTE_ID,
      TEMP_MIN,
      TEMP_MAX,
    ]);

    // Test exact boundary values - should NOT compromise
    await lote.write.registrarTemperatura([TEMP_MIN, TEMP_MAX]); // Exactly 2-8
    assert.equal(await lote.read.comprometido(), false);

    // Deploy new contract for next test
    const lote2 = await viem.deployContract("LoteTracing", [
      LOTE_ID + "-2",
      TEMP_MIN,
      TEMP_MAX,
    ]);

    // Test just outside boundaries - should compromise (tempMin = 1 < 2)
    await lote2.write.registrarTemperatura([1, TEMP_MAX]); // 1-8, tempMin=1 < 2
    assert.equal(await lote2.read.comprometido(), true);

    // Deploy new contract for next test
    const lote3 = await viem.deployContract("LoteTracing", [
      LOTE_ID + "-3",
      TEMP_MIN,
      TEMP_MAX,
    ]);

    // Test just outside boundaries - should compromise (tempMax = 9 > 8)
    await lote3.write.registrarTemperatura([TEMP_MIN, 9]); // 2-9, tempMax=9 > 8
    assert.equal(await lote3.read.comprometido(), true);
  });

  it("Should create new lot with different parameters", async function () {
    const lote = await viem.deployContract("LoteTracing", [
      LOTE_ID,
      TEMP_MIN,
      TEMP_MAX,
    ]);

    // Verify initial state
    assert.equal(await lote.read.loteId(), LOTE_ID);
    assert.equal(await lote.read.temperaturaMinima(), TEMP_MIN);
    assert.equal(await lote.read.temperaturaMaxima(), TEMP_MAX);

    // Create new lot with different parameters
    const NEW_LOTE_ID = "LOT-2024-002";
    const NEW_TEMP_MIN = -5;
    const NEW_TEMP_MAX = 15;

    await lote.write.crearNuevoLote([NEW_LOTE_ID, NEW_TEMP_MIN, NEW_TEMP_MAX]);

    // Verify new state
    assert.equal(await lote.read.loteId(), NEW_LOTE_ID);
    assert.equal(await lote.read.temperaturaMinima(), NEW_TEMP_MIN);
    assert.equal(await lote.read.temperaturaMaxima(), NEW_TEMP_MAX);
    assert.equal(await lote.read.comprometido(), false);
    assert.equal(await lote.read.tempRegMinima(), 0);
    assert.equal(await lote.read.tempRegMaxima(), 0);

    // Verify owner is reset to caller
    const propietarioActual = await lote.read.propietarioActual();
    assert.equal(
      propietarioActual.toLowerCase(),
      fabricante.account.address.toLowerCase()
    );
  });

  it("Should allow anyone to create new lot", async function () {
    const lote = await viem.deployContract("LoteTracing", [
      LOTE_ID,
      TEMP_MIN,
      TEMP_MAX,
    ]);

    // Transfer custody to distributor first
    await lote.write.transferirCustodia([distribuidor.account.address]);

    // Distributor creates new lot
    const NEW_LOTE_ID = "LOT-DIST-001";
    await distribuidor.writeContract({
      address: lote.address,
      abi: lote.abi,
      functionName: "crearNuevoLote",
      args: [NEW_LOTE_ID, -10, 20],
    });

    // Verify distributor is now the owner
    const propietarioActual = await lote.read.propietarioActual();
    assert.equal(
      propietarioActual.toLowerCase(),
      distribuidor.account.address.toLowerCase()
    );
    assert.equal(await lote.read.loteId(), NEW_LOTE_ID);
  });
});
