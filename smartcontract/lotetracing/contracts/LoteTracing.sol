
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

/**
 * @title LoteDeProductoTrazablePoC
 * @author AI Assistant for MediSupply
 * @notice Prueba de Concepto para la trazabilidad de un lote de producto,
 * centrada en la integridad de la cadena de frío y la transferencia de custodia.
 */
contract LoteDeProductoTrazablePoC {

    //==============================================================
    // VARIABLES DE ESTADO
    //==============================================================

    // --- Inmutables (Acta de Nacimiento del Lote) ---
    string public loteId;
    address public immutable fabricante;
    int8 public immutable temperaturaMinima;
    int8 public immutable temperaturaMaxima;

    // --- Dinámicas (Estado Actual) ---
    address public propietarioActual;
    bool public comprometido; // Simplificación: true si la cadena de frío se rompió

    //==============================================================
    // EVENTOS (El historial inmutable)
    //==============================================================

    event LoteCreado(string indexed loteId, address indexed fabricante, int8 temperaturaMinima, int8 temperaturaMaxima);
    event CustodiaTransferida(address indexed propietarioAnterior, address indexed nuevoPropietario);
    event LoteComprometido(int8 temperaturaRegistrada, string motivo);

    //==============================================================
    // MODIFICADOR DE ACCESO
    //==============================================================

    modifier soloPropietario() {
        require(msg.sender == propietarioActual, "Accion solo permitida para el propietario actual");
        _;
    }

    //==============================================================
    // CONSTRUCTOR
    //==============================================================

    constructor(
        string memory _loteId,
        int8 _tempMin,
        int8 _tempMax
    ) {
        loteId = _loteId;
        fabricante = msg.sender;
        propietarioActual = msg.sender; // El fabricante es el primer propietario
        temperaturaMinima = _tempMin;
        temperaturaMaxima = _tempMax;
        comprometido = false;

        emit LoteCreado(_loteId, fabricante, _tempMin, _tempMax);
    }

    //==============================================================
    // FUNCIONES PRINCIPALES
    //==============================================================

    /**
     * @notice Registra una lectura de temperatura. Solo el propietario actual puede hacerlo.
     * Si la temperatura está fuera de rango, el lote se marca como comprometido.
     */
    function registrarTemperatura(int8 _temperatura) external soloPropietario {
        require(!comprometido, "El lote ya esta comprometido");

        if (_temperatura < temperaturaMinima || _temperatura > temperaturaMaxima) {
            comprometido = true;
            emit LoteComprometido(_temperatura, "Temperatura fuera de rango");
        }
    }

    /**
     * @notice Transfiere la propiedad y responsabilidad del lote a un nuevo custodio.
     */
    function transferirCustodia(address _nuevoPropietario) external soloPropietario {
        require(_nuevoPropietario != address(0), "Direccion invalida");
        
        address propietarioAnterior = propietarioActual;
        propietarioActual = _nuevoPropietario;

        emit CustodiaTransferida(propietarioAnterior, _nuevoPropietario);
    }
}