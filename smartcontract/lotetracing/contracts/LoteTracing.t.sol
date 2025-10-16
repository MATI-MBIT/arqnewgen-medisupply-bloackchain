// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {LoteDeProductoTrazable} from "./LoteTracing.sol";
import {Test} from "forge-std/Test.sol";

contract LoteTracingTest is Test {
    LoteDeProductoTrazable lote;
    
    address fabricante = address(0x1);
    address sensor1 = address(0x2);
    address sensor2 = address(0x3);
    address distribuidor = address(0x4);
    address farmacia = address(0x5);
    
    string constant SKU = "MED-001";
    string constant LOTE_ID = "LOT-2024-001";
    uint256 fechaVencimiento;
    int8 constant TEMP_MIN = 2;
    int8 constant TEMP_MAX = 8;

    function setUp() public {
        fechaVencimiento = block.timestamp + 365 days;
        
        vm.prank(fabricante);
        lote = new LoteDeProductoTrazable(
            SKU,
            LOTE_ID,
            fechaVencimiento,
            TEMP_MIN,
            TEMP_MAX
        );
    }

    function test_InitialValues() public view {
        assertEq(lote.sku(), SKU);
        assertEq(lote.loteId(), LOTE_ID);
        assertEq(lote.fabricante(), fabricante);
        assertEq(lote.propietarioActual(), fabricante);
        assertEq(lote.fechaVencimiento(), fechaVencimiento);
        assertEq(lote.temperaturaMinima(), TEMP_MIN);
        assertEq(lote.temperaturaMaxima(), TEMP_MAX);
        assertEq(uint(lote.estado()), uint(LoteDeProductoTrazable.EstadoLote.Creado));
    }

    function test_GestionarSensor() public {
        vm.prank(fabricante);
        lote.gestionarSensor(sensor1, true);
        
        assertTrue(lote.sensoresAutorizados(sensor1));
        
        vm.prank(fabricante);
        lote.gestionarSensor(sensor1, false);
        
        assertFalse(lote.sensoresAutorizados(sensor1));
    }

    function test_GestionarSensor_SoloFabricante() public {
        vm.prank(distribuidor);
        vm.expectRevert("Solo el fabricante puede realizar esta accion");
        lote.gestionarSensor(sensor1, true);
    }

    function test_RegistrarTemperatura() public {
        // Autorizar sensor
        vm.prank(fabricante);
        lote.gestionarSensor(sensor1, true);
        
        // Registrar temperatura válida
        vm.prank(sensor1);
        lote.registrarTemperatura(5);
        
        LoteDeProductoTrazable.LecturaTemperatura[] memory lecturas = lote.obtenerLecturasTemperatura();
        assertEq(lecturas.length, 1);
        assertEq(lecturas[0].temperatura, 5);
        assertEq(lecturas[0].idSensor, sensor1);
    }

    function test_RegistrarTemperatura_SensorNoAutorizado() public {
        vm.prank(sensor1);
        vm.expectRevert("El sensor no esta autorizado");
        lote.registrarTemperatura(5);
    }

    function test_RegistrarTemperatura_FueraDeRango() public {
        // Autorizar sensor
        vm.prank(fabricante);
        lote.gestionarSensor(sensor1, true);
        
        // Registrar temperatura fuera de rango
        vm.prank(sensor1);
        lote.registrarTemperatura(15); // Mayor que TEMP_MAX (8)
        
        // Verificar que el lote se marcó como comprometido
        assertEq(uint(lote.estado()), uint(LoteDeProductoTrazable.EstadoLote.Comprometido));
    }

    function test_TransferirCustodia() public {
        vm.prank(fabricante);
        lote.transferirCustodia(distribuidor);
        
        assertEq(lote.propietarioActual(), distribuidor);
        assertEq(uint(lote.estado()), uint(LoteDeProductoTrazable.EstadoLote.EnTransito));
        
        // Verificar historial
        LoteDeProductoTrazable.HistorialCustodia[] memory historial = lote.obtenerHistorialCustodia();
        assertEq(historial.length, 2);
        assertEq(historial[1].propietario, distribuidor);
    }

    function test_TransferirCustodia_SoloPropietario() public {
        vm.prank(distribuidor);
        vm.expectRevert("Solo el propietario actual puede realizar esta accion");
        lote.transferirCustodia(farmacia);
    }

    function test_TransferirCustodia_DireccionInvalida() public {
        vm.prank(fabricante);
        vm.expectRevert("Direccion de propietario invalida");
        lote.transferirCustodia(address(0));
    }

    function test_CicloCompleto() public {
        // 1. Autorizar sensor
        vm.prank(fabricante);
        lote.gestionarSensor(sensor1, true);
        
        // 2. Registrar temperaturas válidas
        vm.prank(sensor1);
        lote.registrarTemperatura(4);
        
        vm.prank(sensor1);
        lote.registrarTemperatura(6);
        
        // 3. Transferir a distribuidor
        vm.prank(fabricante);
        lote.transferirCustodia(distribuidor);
        
        // 4. Transferir a farmacia
        vm.prank(distribuidor);
        lote.transferirCustodia(farmacia);
        
        // Verificaciones finales
        assertEq(lote.propietarioActual(), farmacia);
        assertEq(uint(lote.estado()), uint(LoteDeProductoTrazable.EstadoLote.EnAlmacen));
        
        LoteDeProductoTrazable.LecturaTemperatura[] memory lecturas = lote.obtenerLecturasTemperatura();
        assertEq(lecturas.length, 2);
        
        LoteDeProductoTrazable.HistorialCustodia[] memory historial = lote.obtenerHistorialCustodia();
        assertEq(historial.length, 3);
    }

    function testFuzz_RegistrarTemperatura(int8 temperatura) public {
        vm.assume(temperatura >= -50 && temperatura <= 50);
        
        // Autorizar sensor
        vm.prank(fabricante);
        lote.gestionarSensor(sensor1, true);
        
        // Registrar temperatura
        vm.prank(sensor1);
        lote.registrarTemperatura(temperatura);
        
        // Verificar estado según temperatura
        if (temperatura < TEMP_MIN || temperatura > TEMP_MAX) {
            assertEq(uint(lote.estado()), uint(LoteDeProductoTrazable.EstadoLote.Comprometido));
        } else {
            assertEq(uint(lote.estado()), uint(LoteDeProductoTrazable.EstadoLote.Creado));
        }
    }
}