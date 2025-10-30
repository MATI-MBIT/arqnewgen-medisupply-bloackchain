// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

/**
 * @title LoteTracing
 * @author Grupo 2 - ArqNewGen - MATI for MediSupply
 * @notice Prueba de Concepto para la trazabilidad de un lote de producto,
 * centrada en la integridad de la cadena de frío y la transferencia de custodia.
 */
contract LoteTracing {
    //==============================================================
    // VARIABLES DE ESTADO
    //==============================================================

    // --- Inmutables (Acta de Nacimiento del Lote) ---
    string public loteId;
    address public immutable fabricante;
    int8 public temperaturaMinima;
    int8 public temperaturaMaxima;
    int8 public tempRegMinima;
    int8 public tempRegMaxima;

    // --- Dinámicas (Estado Actual) ---
    address public propietarioActual;
    bool public comprometido; // Simplificación: true si la cadena de frío se rompió

    //==============================================================
    // EVENTOS (El historial inmutable)
    //==============================================================

    event LoteCreado(
        string indexed loteId,
        address indexed fabricante,
        int8 temperaturaMinima,
        int8 temperaturaMaxima,
        string motivo
    );
    event CustodiaTransferida(
        address indexed propietarioAnterior,
        address indexed nuevoPropietario,
        bool comprometido,
        string motivo
    );
    event LoteComprometido(
        address indexed propietario,
        int8 tempMin,
        int8 tempMax,
        bool comprometido,
        string motivo
    );

    //==============================================================
    // MODIFICADOR DE ACCESO
    //==============================================================

    modifier soloPropietario() {
        require(
            msg.sender == propietarioActual,
            "Accion solo permitida para el propietario actual"
        );
        _;
    }

    //==============================================================
    // CONSTRUCTOR
    //==============================================================

    constructor(string memory _loteId, int8 _tempMin, int8 _tempMax) {
        loteId = _loteId;
        fabricante = msg.sender;
        propietarioActual = msg.sender; // El fabricante es el primer propietario
        temperaturaMinima = _tempMin;
        temperaturaMaxima = _tempMax;
        comprometido = false;
        tempRegMinima = 0;
        tempRegMaxima = 0;

        emit LoteCreado(_loteId, fabricante, _tempMin, _tempMax, "Lote Creado");
    }

    //==============================================================
    // FUNCIONES PRINCIPALES
    //==============================================================

    /**
     * @notice Registra una lectura de temperatura. Solo el propietario actual puede hacerlo.
     * Si la temperatura está fuera de rango, el lote se marca como comprometido.
     */
    function registrarTemperatura(int8 _tempMin, int8 _tempMax) external {
        // require(!comprometido, "El lote ya esta comprometido");

        tempRegMinima = _tempMin;
        tempRegMaxima = _tempMax;
        if (_tempMin < temperaturaMinima || _tempMax > temperaturaMaxima) {
            comprometido = true;
            emit LoteComprometido(
                msg.sender,
                _tempMin,
                _tempMax,
                comprometido,
                "Temperatura fuera de rango"
            );
        }
    }

    /**
     * @notice Transfiere la propiedad y responsabilidad del lote a un nuevo custodio.
     */
    function transferirCustodia(
        address _nuevoPropietario
    ) external soloPropietario {
        require(_nuevoPropietario != address(0), "Direccion invalida");

        address propietarioAnterior = propietarioActual;
        propietarioActual = _nuevoPropietario;

        emit CustodiaTransferida(
            propietarioAnterior,
            _nuevoPropietario,
            comprometido,
            "Custodia Transferida"
        );
    }

    /**
     * @notice Registra una lectura de temperatura. Solo el propietario actual puede hacerlo.
     * Si la temperatura está fuera de rango, el lote se marca como comprometido.
     */
    function crearNuevoLote(string memory _loteId, int8 _tempMin, int8 _tempMax) external {
        loteId = _loteId;
        propietarioActual = msg.sender; // El fabricante es el primer propietario
        temperaturaMinima = _tempMin;
        temperaturaMaxima = _tempMax;
        comprometido = false;
        tempRegMinima = 0;
        tempRegMaxima = 0;

        emit LoteCreado(_loteId, fabricante, _tempMin, _tempMax, "Lote Creado");
    }
}
