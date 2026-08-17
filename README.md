Ingestão de Telemetria IoT & GPS em Alta Frequência (graaa-golang-telemetry-api).

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
