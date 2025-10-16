import assert from "node:assert/strict";
import { describe, it } from "node:test";

import { network } from "hardhat";

describe("LoteTracing", async function () {
  const { viem } = await network.connect();
  const publicClient = await viem.getPublicClient();

  // Test addresses
  const [fabricante, sensor1, distribuidor, farmacia] = await viem.getWalletClients();

  // Contract parameters
  const SKU = "MED-001";
  const LOTE_ID = "LOT-2024-001";
  const TEMP_MIN = 2;
  const TEMP_MAX = 8;

  it("Should deploy and initialize correctly", async function () {
    const fechaVencimiento = BigInt(Math.floor(Date.now() / 1000) + 365 * 24 * 60 * 60);
    
    const lote = await viem.deployContract("LoteDeProductoTrazable", [
      SKU,
      LOTE_ID,
      fechaVencimiento,
      TEMP_MIN,
      TEMP_MAX
    ]);

    const sku = await lote.read.sku();
    const loteId = await lote.read.loteId();
    const fabricanteAddr = await lote.read.fabricante();
    const propietarioActual = await lote.read.propietarioActual();

    assert.equal(sku, SKU);
    assert.equal(loteId, LOTE_ID);
    assert.equal(fabricanteAddr.toLowerCase(), fabricante.account.address.toLowerCase());
    assert.equal(propietarioActual.toLowerCase(), fabricante.account.address.toLowerCase());
  });

  it("Should manage sensors correctly", async function () {
    const fechaVencimiento = BigInt(Math.floor(Date.now() / 1000) + 365 * 24 * 60 * 60);
    
    const lote = await viem.deployContract("LoteDeProductoTrazable", [
      SKU,
      LOTE_ID,
      fechaVencimiento,
      TEMP_MIN,
      TEMP_MAX
    ]);

    // Authorize sensor
    await lote.write.gestionarSensor([sensor1.account.address, true]);
    
    const isAuthorized = await lote.read.sensoresAutorizados([sensor1.account.address]);
    assert.equal(isAuthorized, true);

    // Deauthorize sensor
    await lote.write.gestionarSensor([sensor1.account.address, false]);
    
    const isStillAuthorized = await lote.read.sensoresAutorizados([sensor1.account.address]);
    assert.equal(isStillAuthorized, false);
  });

  it("Should register temperature correctly", async function () {
    const fechaVencimiento = BigInt(Math.floor(Date.now() / 1000) + 365 * 24 * 60 * 60);
    
    const lote = await viem.deployContract("LoteDeProductoTrazable", [
      SKU,
      LOTE_ID,
      fechaVencimiento,
      TEMP_MIN,
      TEMP_MAX
    ]);

    // Authorize sensor (using fabricante client since they deployed the contract)
    await lote.write.gestionarSensor([sensor1.account.address, true]);

    // Register temperature (using sensor1 client)
    const temperatura = 5;
    await sensor1.writeContract({
      address: lote.address,
      abi: lote.abi,
      functionName: "registrarTemperatura",
      args: [temperatura]
    });

    const lecturas = await lote.read.obtenerLecturasTemperatura() as any[];
    assert.equal(lecturas.length, 1);
    assert.equal(lecturas[0].temperatura, temperatura);
    assert.equal(lecturas[0].idSensor.toLowerCase(), sensor1.account.address.toLowerCase());
  });

  it("Should mark lot as compromised when temperature is out of range", async function () {
    const fechaVencimiento = BigInt(Math.floor(Date.now() / 1000) + 365 * 24 * 60 * 60);
    
    const lote = await viem.deployContract("LoteDeProductoTrazable", [
      SKU,
      LOTE_ID,
      fechaVencimiento,
      TEMP_MIN,
      TEMP_MAX
    ]);

    // Authorize sensor
    await lote.write.gestionarSensor([sensor1.account.address, true]);

    // Register temperature out of range (using sensor1 client)
    const temperaturaAlta = 15; // Above TEMP_MAX (8)
    await sensor1.writeContract({
      address: lote.address,
      abi: lote.abi,
      functionName: "registrarTemperatura",
      args: [temperaturaAlta]
    });

    const estado = await lote.read.estado();
    assert.equal(estado, 3); // EstadoLote.Comprometido
  });

  it("Should transfer custody correctly", async function () {
    const fechaVencimiento = BigInt(Math.floor(Date.now() / 1000) + 365 * 24 * 60 * 60);
    
    const lote = await viem.deployContract("LoteDeProductoTrazable", [
      SKU,
      LOTE_ID,
      fechaVencimiento,
      TEMP_MIN,
      TEMP_MAX
    ]);

    // Transfer to distributor
    await lote.write.transferirCustodia([distribuidor.account.address]);

    const nuevoPropietario = await lote.read.propietarioActual();
    const estado = await lote.read.estado();
    
    assert.equal(nuevoPropietario.toLowerCase(), distribuidor.account.address.toLowerCase());
    assert.equal(estado, 1); // EstadoLote.EnTransito

    // Check custody history
    const historial = await lote.read.obtenerHistorialCustodia() as any[];
    assert.equal(historial.length, 2);
    assert.equal(historial[1].propietario.toLowerCase(), distribuidor.account.address.toLowerCase());
  });

  it("Should complete full traceability cycle", async function () {
    const fechaVencimiento = BigInt(Math.floor(Date.now() / 1000) + 365 * 24 * 60 * 60);
    const deploymentBlockNumber = await publicClient.getBlockNumber();
    
    const lote = await viem.deployContract("LoteDeProductoTrazable", [
      SKU,
      LOTE_ID,
      fechaVencimiento,
      TEMP_MIN,
      TEMP_MAX
    ]);

    // 1. Authorize sensor
    await lote.write.gestionarSensor([sensor1.account.address, true]);

    // 2. Register multiple temperature readings (using sensor1 client)
    const temperaturas = [4, 5, 6, 7];
    for (const temp of temperaturas) {
      await sensor1.writeContract({
        address: lote.address,
        abi: lote.abi,
        functionName: "registrarTemperatura",
        args: [temp]
      });
    }

    // 3. Transfer custody: Fabricante -> Distribuidor -> Farmacia
    await lote.write.transferirCustodia([distribuidor.account.address]);
    await distribuidor.writeContract({
      address: lote.address,
      abi: lote.abi,
      functionName: "transferirCustodia",
      args: [farmacia.account.address]
    });

    // Verify final state
    const propietarioFinal = await lote.read.propietarioActual();
    const estadoFinal = await lote.read.estado();
    const lecturas = await lote.read.obtenerLecturasTemperatura() as any[];
    const historial = await lote.read.obtenerHistorialCustodia() as any[];

    assert.equal(propietarioFinal.toLowerCase(), farmacia.account.address.toLowerCase());
    assert.equal(estadoFinal, 2); // EstadoLote.EnAlmacen
    assert.equal(lecturas.length, temperaturas.length);
    assert.equal(historial.length, 3); // Fabricante, Distribuidor, Farmacia

    // Verify all temperature events were emitted
    const events = await publicClient.getContractEvents({
      address: lote.address,
      abi: lote.abi,
      eventName: "TemperaturaRegistrada",
      fromBlock: deploymentBlockNumber,
      strict: true,
    });

    assert.equal(events.length, temperaturas.length);
  });

  it("Should reject unauthorized operations", async function () {
    const fechaVencimiento = BigInt(Math.floor(Date.now() / 1000) + 365 * 24 * 60 * 60);
    
    const lote = await viem.deployContract("LoteDeProductoTrazable", [
      SKU,
      LOTE_ID,
      fechaVencimiento,
      TEMP_MIN,
      TEMP_MAX
    ]);

    // Try to register temperature from unauthorized sensor
    await assert.rejects(
      sensor1.writeContract({
        address: lote.address,
        abi: lote.abi,
        functionName: "registrarTemperatura",
        args: [5]
      }),
      /El sensor no esta autorizado/
    );
  });
});