# AntalyaKart Telegram advisor

`antalyakart-advisor` is the Telegram bot `@antalyakart_advisor_bot`. It answers questions about public buses in Antalya with a Google ADK for Go agent (`LlmAgent` and `Runner`, at most six model rounds) and the AntalyaKart transit tools. After deploy, Telegram posts updates to `https://mcp.goturkey.club/antalyakart-advisor/webhook`.

Transit data comes from the same API as [online.antalyakart.com.tr](https://online.antalyakart.com.tr/#/home).

Chat history is one JSON file, a map of chat id to messages. The default name is `chat-history.json` next to the executable, so the process can be started from another directory. `CHAT_HISTORY_PATH` overrides it; a relative value is still resolved from the executable directory.

## Run

```bash
go build -o bin/antalyakart-advisor ./cmd/antalyakart-advisor
```

The binary listens on `127.0.0.1:8091`. `GET /healthz` returns `ok`.

| Variable | Required | Default |
|---|---|---|
| `TELEGRAM_BOT_TOKEN` | yes | |
| `GOOGLE_API_KEY` | yes | Gemini / ADK key |
| `GEMINI_MODEL` | no | `gemini-flash-latest` |
| `WEBHOOK_URL` | no | empty skips webhook registration |
| `TELEGRAM_WEBHOOK_SECRET` | no | empty accepts any caller |
| `BOT_ADDR` | no | `127.0.0.1:8091` |
| `BOT_PATH` | no | `/antalyakart-advisor` |
| `CHAT_HISTORY_PATH` | no | `chat-history.json` beside the executable |
| `ANTALYAKART_BASE_URL` | no | `https://service.kentkart.com/rl1` |
| `ANTALYAKART_REGION` | no | `026` |
| `ANTALYAKART_LANG` | no | `tr` |
| `ANTALYAKART_AUTH_TYPE` | no | `4` |

Put the token and API key in a gitignored env file or in the server env. `.env.example` lists the names. Deploy writes `/var/www/mcp/antalyakart-advisor.env` from GitHub Actions secrets `TELEGRAM_BOT_TOKEN` and `GOOGLE_API_KEY`, using SSH secrets `TR_SSH_KEY`, `TR_SSH_USER`, and `TR_SSH_PORT` (`.github/workflows/deploy-advisor.yml`).

## Tools

- `search_routes_and_stops` — routes, stops, and places by keyword
- `nearby_places_stops_and_kiosks` — places, stops, and card top-up points near a coordinate
- `nearest_buses_for_stop` — buses approaching a stop
- `route_path_and_vehicles` — route geometry, stops, schedule, or live vehicles only
- `plan_direct_trip_between_stops` — direct buses between two stop searches
- `summarize_stop_arrivals` — upcoming arrivals at a stop, grouped by route
