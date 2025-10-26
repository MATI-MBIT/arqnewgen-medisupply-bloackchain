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
        // Registrar rango de temperatura válido (cualquiera puede hacerlo)
        lote.registrarTemperatura(TEMP_MIN, TEMP_MAX);
        
        // Verificar que el lote no está comprometido
        assertEq(lote.comprometido(), false);
    }

    function test_RegistrarTemperatura_CualquieraPuede() public {
        // Ahora cualquiera puede registrar temperaturas
        vm.prank(distribuidor);
        lote.registrarTemperatura(TEMP_MIN, TEMP_MAX);
        
        assertEq(lote.comprometido(), false);
    }

    function test_RegistrarTemperatura_RangoInvalido() public {
        // Registrar rango que no incluye las temperaturas del contrato
        lote.registrarTemperatura(10, 15); // Rango 10-15 no incluye 2-8
        
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
        // 1. Registrar rangos de temperatura válidos
        lote.registrarTemperatura(TEMP_MIN, TEMP_MAX);
        lote.registrarTemperatura(0, 10);
        
        // 2. Transferir a distribuidor
        vm.prank(fabricante);
        lote.transferirCustodia(distribuidor);
        
        // 3. Distribuidor registra rango de temperatura
        lote.registrarTemperatura(TEMP_MIN, TEMP_MAX);
        
        // 4. Transferir a farmacia
        vm.prank(distribuidor);
        lote.transferirCustodia(farmacia);
        
        // Verificaciones finales
        assertEq(lote.propietarioActual(), farmacia);
        assertEq(lote.comprometido(), false);
    }

    function test_LoteComprometidoNoPermiteTemperaturas() public {
        // Comprometer el lote con rango inválido
        lote.registrarTemperatura(10, 15); // Rango no incluye 2-8
        
        assertTrue(lote.comprometido());
        
        // Intentar registrar otro rango
        vm.expectRevert("El lote ya esta comprometido");
        lote.registrarTemperatura(TEMP_MIN, TEMP_MAX);
    }

    function testFuzz_RegistrarTemperatura(int8 tempMin, int8 tempMax) public {
        vm.assume(tempMin >= -50 && tempMax <= 50 && tempMin <= tempMax);
        
        // Registrar rango de temperatura
        lote.registrarTemperatura(tempMin, tempMax);
        
        // Verificar estado según si el rango incluye las temperaturas del contrato
        if (TEMP_MIN < tempMin || TEMP_MIN > tempMax || TEMP_MAX < tempMin || TEMP_MAX > tempMax) {
            assertTrue(lote.comprometido());
        } else {
            assertFalse(lote.comprometido());
        }
    }
}