Ingestão de Telemetria IoT & GPS em Alta Frequência (graaa-golang-telemetry-api).

## Configuração de segurança

- `TELEMETRY_DEVICE_KEYS` é obrigatório e deve ser um objeto JSON com uma chave aleatória (mínimo 32 bytes) por IMEI, por exemplo `{"IMEI_EXEMPLO":"substituir-por-segredo-aleatorio-de-32-bytes"}`. Configure os rastreadores para enviar o IMEI em `X-Tracker-IMEI` e a respetiva chave em `X-Tracker-Token`.
- Cada rastreador só pode enviar pacotes para a viatura cujo `tracker_imei` corresponde à credencial. O `vehicleId` no corpo, se fornecido, tem de coincidir com essa viatura. A ingestão em lote só aceita pacotes do rastreador autenticado.
- Defina `JWT_SECRET` igual ao auth API e com pelo menos 32 bytes para consultar histórico. Defina `FLEET_TELEMETRY_TOKEN` igual ao valor configurado na fleet API para publicar atualizações internas.
- Defina `ALLOWED_ORIGINS` para as origens exatas da aplicação frontend; sem configuração, apenas `http://localhost:5173` é autorizado.
- Não reutilize chaves entre dispositivos nem coloque segredos de dispositivo no repositório; use o gestor de segredos do ambiente de execução.

Esta API é o coração do rastreamento da frota:

    Alta Performance & Baixa Latência: Recebe pacotes de GPS a cada 2 a 5 segundos de centenas de táxis e motos.

    Compatibilidade Multi-Protocolo: Suporta tanto requisições HTTP REST/JSON de aplicativos móveis de motoristas quanto pacotes diretos de Rastreadores GPS físicos (IMEI/OBD-II/Teltonika/GT06).

    Buffer e Cache no Redis: Grava a posição instantânea no Redis em microssegundos e despacha eventos para a fleet-api e WebSockets sem engarrafar o banco relacional.

    Motor de Regras de Alerta Automático: Detecta instantaneamente excesso de velocidade (>80 km/h na Marginal/Via Expressa), nível crítico de combustível (<15%) ou superaquecimento do motor (>98°C).

📁 Estrutura da Pasta: backend-telemetry-golang
code Text

backend-telemetry-golang/
├── .env
├── go.mod
├── main.go
├── config/
│   ├── database.go          # Conexão PostgreSQL (fleet_db)
│   └── redis.go             # Conexão Redis para cache de coordenadas GPS em alta frequência
├── models/
│   └── telemetry_models.go  # Structs de Posições, OBD-II e Regras de Alerta
├── services/
│   ├── alert_engine.go      # Motor de deteção de excesso de velocidade e motor quente
│   └── redis_tracker.go     # Gravação em tempo real no Redis
└── handlers/
    ├── ingest_handler.go    # Ingestão em lote e individual de pacotes GPS/OBD
    └── history_handler.go   # Consulta de rotas e histórico de telemetria por período
