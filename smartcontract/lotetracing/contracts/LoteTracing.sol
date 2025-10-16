// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

/**
 * @title LoteDeProductoTrazable
 * @author AI Assistant for MediSupply
 * @notice Contrato para la trazabilidad de un lote de producto farmacéutico sensible,
 * gestionando su ciclo de vida, cadena de frío y custodia.
 */
contract LoteDeProductoTrazable {
    //==============================================================
    // ESTRUCTURAS DE DATOS Y ESTADOS
    //==============================================================

    enum EstadoLote {
        Creado,
        EnTransito,
        EnAlmacen,
        Comprometido,
        Entregado
    }

    struct HistorialCustodia {
        address propietario;
        uint256 timestamp;
    }

    struct LecturaTemperatura {
        int8 temperatura;
        uint256 timestamp;
        address idSensor;
    }

    //==============================================================
    // VARIABLES DE ESTADO
    //==============================================================

    // --- Datos Inmutables ---
    string public sku;
    string public loteId;
    address public immutable fabricante;
    uint256 public immutable fechaFabricacion;
    uint256 public immutable fechaVencimiento;
    int8 public immutable temperaturaMinima;
    int8 public immutable temperaturaMaxima;

    // --- Datos Dinámicos ---
    address public propietarioActual;
    EstadoLote public estado;
    mapping(address => bool) public sensoresAutorizados;

    HistorialCustodia[] private _historialCustodia;
    LecturaTemperatura[] private _lecturasTemperatura;

    //==============================================================
    // EVENTOS
    //==============================================================

    event LoteCreado(
        string indexed sku,
        string indexed loteId,
        address indexed fabricante
    );
    event CustodiaTransferida(
        address indexed propietarioAnterior,
        address indexed nuevoPropietario,
        uint256 timestamp
    );
    event TemperaturaRegistrada(
        string indexed loteId,
        int8 temperatura,
        uint256 timestamp
    );
    event LoteComprometido(
        string indexed loteId,
        int8 temperaturaRegistrada,
        string motivo
    );

    //==============================================================
    // MODIFICADORES (REGLAS DE ACCESO)
    //==============================================================

    modifier soloFabricante() {
        require(
            msg.sender == fabricante,
            "Solo el fabricante puede realizar esta accion"
        );
        _;
    }

    modifier soloPropietario() {
        require(
            msg.sender == propietarioActual,
            "Solo el propietario actual puede realizar esta accion"
        );
        _;
    }

    modifier soloSensorAutorizado() {
        require(
            sensoresAutorizados[msg.sender],
            "El sensor no esta autorizado"
        );
        _;
    }

    //==============================================================
    // CONSTRUCTOR
    //==============================================================

    constructor(
        string memory _sku,
        string memory _loteId,
        uint256 _fechaVencimiento,
        int8 _tempMin,
        int8 _tempMax
    ) {
        sku = _sku;
        loteId = _loteId;
        fabricante = msg.sender;
        propietarioActual = msg.sender;
        fechaFabricacion = block.timestamp;
        fechaVencimiento = _fechaVencimiento;
        temperaturaMinima = _tempMin;
        temperaturaMaxima = _tempMax;
        estado = EstadoLote.Creado;

        _historialCustodia.push(
            HistorialCustodia({
                propietario: msg.sender,
                timestamp: block.timestamp
            })
        );

        emit LoteCreado(_sku, _loteId, msg.sender);
    }

    //==============================================================
    // FUNCIONES PRINCIPALES
    //==============================================================

    /**
     * @notice Permite al fabricante autorizar o desautorizar un sensor de IoT.
     * @param _sensor La dirección del wallet del sensor.
     * @param _autorizado El estado de autorización (true o false).
     */
    function gestionarSensor(
        address _sensor,
        bool _autorizado
    ) external soloFabricante {
        sensoresAutorizados[_sensor] = _autorizado;
    }

    /**
     * @notice Registra una nueva lectura de temperatura desde un sensor autorizado.
     * @param _temperatura La temperatura actual del lote.
     */
    function registrarTemperatura(
        int8 _temperatura
    ) external soloSensorAutorizado {
        require(
            estado != EstadoLote.Comprometido,
            "El lote ya esta comprometido"
        );

        _lecturasTemperatura.push(
            LecturaTemperatura({
                temperatura: _temperatura,
                timestamp: block.timestamp,
                idSensor: msg.sender
            })
        );

        emit TemperaturaRegistrada(loteId, _temperatura, block.timestamp);

        if (
            _temperatura < temperaturaMinima || _temperatura > temperaturaMaxima
        ) {
            _marcarComoComprometido(_temperatura);
        }
    }

    /**
     * @notice Transfiere la custodia del lote a un nuevo propietario.
     * @param _nuevoPropietario La dirección del wallet del nuevo custodio.
     */
    function transferirCustodia(
        address _nuevoPropietario
    ) external soloPropietario {
        require(
            _nuevoPropietario != address(0),
            "Direccion de propietario invalida"
        );

        address propietarioAnterior = propietarioActual;
        propietarioActual = _nuevoPropietario;

        _historialCustodia.push(
            HistorialCustodia({
                propietario: _nuevoPropietario,
                timestamp: block.timestamp
            })
        );

        // Lógica simple para actualizar el estado basado en la transferencia
        if (estado == EstadoLote.Creado) {
            estado = EstadoLote.EnTransito;
        } else if (estado == EstadoLote.EnTransito) {
            // Se asume que la segunda transferencia es al almacén
            estado = EstadoLote.EnAlmacen;
        }

        emit CustodiaTransferida(
            propietarioAnterior,
            _nuevoPropietario,
            block.timestamp
        );
    }

    //==============================================================
    // FUNCIONES INTERNAS
    //==============================================================

    function _marcarComoComprometido(int8 _temperatura) private {
        estado = EstadoLote.Comprometido;
        emit LoteComprometido(
            loteId,
            _temperatura,
            "Temperatura fuera de rango"
        );
    }

    //==============================================================
    // FUNCIONES DE SOLO LECTURA
    //==============================================================

    function obtenerHistorialCustodia()
        external
        view
        returns (HistorialCustodia[] memory)
    {
        return _historialCustodia;
    }

    function obtenerLecturasTemperatura()
        external
        view
        returns (LecturaTemperatura[] memory)
    {
        return _lecturasTemperatura;
    }
}
