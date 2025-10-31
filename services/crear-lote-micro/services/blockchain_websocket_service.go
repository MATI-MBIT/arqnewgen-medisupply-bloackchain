package services

import (
	"log"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
)

type BlockchainWebsocketService struct {
	u                 string
	blockchainService *BlockchainService
	damageCaller      *DamageServiceCaller
}

func NewBlockchainWebsocketService(u string, blockchainService *BlockchainService, damageCaller *DamageServiceCaller) *BlockchainWebsocketService {
	return &BlockchainWebsocketService{u: u,
		blockchainService: blockchainService,
		damageCaller:      damageCaller,
	}
}

func (s *BlockchainWebsocketService) StartBlockchainWebsocket(address string) {

	log.Printf("Iniciando a %s", address)

	// Canal para manejar la interrupción (Ctrl+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// 1. Definir la URL del WebSocket
	// Esta es la URL de tu comando wscat
	u := s.u

	log.Printf("Conectando a %s", u)

	// 2. Conectar al servidor WebSocket
	// websocket.DefaultDialer.Dial se conecta a la URL dada.
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatal("Error al conectar:", err)
	}
	defer c.Close()

	log.Println("¡Conectado exitosamente!")

	// 3. Definir el mensaje de suscripción JSON
	// Este es el payload JSON del argumento -x de tu comando
	subscriptionMsg := `{"jsonrpc": "2.0", "method": "eth_subscribe","params": ["alchemy_minedTransactions", {"addresses": [{"to": "` + address + `"}],"includeRemoved": false, "hashesOnly": false}],"id": 1761756245}`

	// 4. Enviar el mensaje de suscripción
	// Enviamos el payload como un mensaje de texto
	err = c.WriteMessage(websocket.TextMessage, []byte(subscriptionMsg))
	if err != nil {
		log.Println("Error al escribir (suscribir):", err)
		return
	}
	log.Println("Suscripción enviada. Esperando mensajes...")

	// Canal para los mensajes recibidos
	done := make(chan struct{})

	// 5. Iniciar una goroutine para leer mensajes del servidor
	go func() {
		defer close(done) // Cierra el canal 'done' cuando esta goroutine termina
		for {
			// Leer mensajes continuamente
			_, message, err := c.ReadMessage()
			if err != nil {
				// Si hay un error (ej. desconexión), lo registramos y salimos
				log.Println("Error al leer:", err)
				return
			}
			// Imprimir el mensaje recibido (la transacción minada)
			log.Printf("Mensaje recibido: %s", message)
			infoLote, err := s.blockchainService.ObtenerInfoLote(address)
			if err != nil {
				log.Println("Error al obtener info del lote:", err)
			} else {
				log.Printf("Info del lote: %v", infoLote)
				if infoLote.Comprometido {
					if err := s.damageCaller.SendLoteInfo(infoLote); err != nil {
						log.Printf("Error enviando LoteInfo al damage service: %v", err)
					}
					log.Printf("❌ Lote comprometido: loteId:%s, contractAddress:%s, propietarioActual:%s, comprometido:%t", infoLote.LoteID, infoLote.ContractAddress, infoLote.PropietarioActual, infoLote.Comprometido)
				} else {
					log.Printf("✅ Lote sin novedades loteId:%s, contractAddress:%s", infoLote.LoteID, infoLote.ContractAddress)
				}
			}
		}
	}()

	// 6. Esperar por una interrupción (Ctrl+C) o por un error de lectura
	select {
	case <-done:
		// Se activó si la goroutine de lectura terminó (ej. por error)
		log.Println("La goroutine de lectura terminó.")
	case <-interrupt:
		// Se activó si el usuario presionó Ctrl+C
		log.Println("Interrupción recibida. Cerrando conexión...")

		// Enviar un mensaje de cierre limpio al servidor
		err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Println("Error al escribir mensaje de cierre:", err)
			return
		}
	}
}
