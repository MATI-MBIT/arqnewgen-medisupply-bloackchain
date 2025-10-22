// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {LoteTracing} from "./LoteTracing.sol";
import {Test} from "forge-std/Test.sol";

contract LoteTracingTest is Test {
    LoteTracing lote;
    
    address fabricante = address(0x1);
    address distribuidor = address(0x2);
    address farmacia = address(0x3);
    
    string constant LOTE_ID = "LOT-2024-001";
    int8 constant TEMP_MIN = 2;
    int8 constant TEMP_MAX = 8;

    function setUp() public {
        vm.prank(fabricante);
        lote = new LoteTracing(
            LOTE_ID,
            TEMP_MIN,
            TEMP_MAX
        );
    }

    function test_InitialValues() public view {
        assertEq(lote.loteId(), LOTE_ID);
        assertEq(lote.fabricante(), fabricante);
        assertEq(lote.propietarioActual(), fabricante);
        assertEq(lote.temperaturaMinima(), TEMP_MIN);
        assertEq(lote.temperaturaMaxima(), TEMP_MAX);
        assertEq(lote.comprometido(), false);
    }

    function test_RegistrarTemperatura_Valida() public {
        // Registrar temperatura válida (solo propietario puede hacerlo)
        vm.prank(fabricante);
        lote.registrarTemperatura(5);
        
        // Verificar que el lote no está comprometido
        assertEq(lote.comprometido(), false);
    }

    function test_RegistrarTemperatura_SoloPropietario() public {
        vm.prank(distribuidor);
        vm.expectRevert("Accion solo permitida para el propietario actual");
        lote.registrarTemperatura(5);
    }

    function test_RegistrarTemperatura_FueraDeRango() public {
        // Registrar temperatura fuera de rango
        vm.prank(fabricante);
        lote.registrarTemperatura(15); // Mayor que TEMP_MAX (8)
        
        // Verificar que el lote se marcó como comprometido
        assertEq(lote.comprometido(), true);
    }

    function test_TransferirCustodia() public {
        vm.prank(fabricante);
        lote.transferirCustodia(distribuidor);
        
        assertEq(lote.propietarioActual(), distribuidor);
    }

    function test_TransferirCustodia_SoloPropietario() public {
        vm.prank(distribuidor);
        vm.expectRevert("Accion solo permitida para el propietario actual");
        lote.transferirCustodia(farmacia);
    }

    function test_TransferirCustodia_DireccionInvalida() public {
        vm.prank(fabricante);
        vm.expectRevert("Direccion invalida");
        lote.transferirCustodia(address(0));
    }

    function test_CicloCompleto() public {
        // 1. Registrar temperaturas válidas
        vm.prank(fabricante);
        lote.registrarTemperatura(4);
        
        vm.prank(fabricante);
        lote.registrarTemperatura(6);
        
        // 2. Transferir a distribuidor
        vm.prank(fabricante);
        lote.transferirCustodia(distribuidor);
        
        // 3. Distribuidor registra temperatura
        vm.prank(distribuidor);
        lote.registrarTemperatura(5);
        
        // 4. Transferir a farmacia
        vm.prank(distribuidor);
        lote.transferirCustodia(farmacia);
        
        // Verificaciones finales
        assertEq(lote.propietarioActual(), farmacia);
        assertEq(lote.comprometido(), false);
    }

    function test_LoteComprometidoNoPermiteTemperaturas() public {
        // Comprometer el lote
        vm.prank(fabricante);
        lote.registrarTemperatura(15); // Fuera de rango
        
        assertTrue(lote.comprometido());
        
        // Intentar registrar otra temperatura
        vm.prank(fabricante);
        vm.expectRevert("El lote ya esta comprometido");
        lote.registrarTemperatura(5);
    }

    function testFuzz_RegistrarTemperatura(int8 temperatura) public {
        vm.assume(temperatura >= -50 && temperatura <= 50);
        
        // Registrar temperatura
        vm.prank(fabricante);
        lote.registrarTemperatura(temperatura);
        
        // Verificar estado según temperatura
        if (temperatura < TEMP_MIN || temperatura > TEMP_MAX) {
            assertTrue(lote.comprometido());
        } else {
            assertFalse(lote.comprometido());
        }
    }
}