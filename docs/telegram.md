# Telegram Bot Conventions

## Structure

- Handlers in `internal/delivery/telegram/handlers/`.
- Router in `internal/delivery/telegram/router.go`.
- Utilities in `internal/delivery/telegram/tgutil/`.

## Patterns

- **FSM** (finite state machine) for multi-step flows (welcome, training creation, joint run creation). State stored in `bot_states` table.
- **Callback data** encoded as `action:payload` pairs via `tgutil.Encode/Decode`.
- **MarkdownV2** escaping via `internal/pkg/tgrescape` — all bot messages must escape special characters.
- Handler methods process a single action or FSM step each.
