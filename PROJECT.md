# Meido Project Guide

## Overview

Meido is a modular Discord bot written in Go and powered by the DiscordGo library. The repository packages a reusable bot framework (`pkg/mio`) alongside the production bot implementation (`internal/meido`) and feature modules under `internal/module`. The bot combines slash-commands, message commands, event-driven workflows, and Postgres-backed persistence to deliver moderation, search, utility, and entertainment features for Discord servers.

## Runtime Architecture

- `cmd/meido/main.go` loads configuration, initialises the Postgres connection, constructs the bot via `internal/meido`, and blocks until shutdown signals arrive.
- `internal/meido.Meido` orchestrates startup: it builds a `pkg/mio/bot.Bot`, registers modules, sets Discord handlers, and exposes lifecycle hooks (`Run`, `Close`).
- `pkg/mio/bot` provides the modular runtime: Discord session management, event fan-out, command dispatch, cooldowns, and an internal event bus.
- Modules implement `pkg/mio/bot.Module`, registering commands, passives, slash commands, and UI handlers during their `Hook` phase. The event system routes Discord events to each active module.
- `internal/database` supplies a Postgres-backed implementation of bot storage interfaces (guild metadata, command logging, processed-event counters, plus module-specific repositories).

## Project Structure and Module Organisation

- `cmd/meido/` – CLI entrypoint; houses `config.json` (sample runtime settings) and `main.go` (process lifecycle management).
- `internal/meido/` – Application orchestration. `meido.go` wires the bot, `logger.go` adapts zap logging, and `logs.go` consumes runtime analytics from the event bus.
- `internal/database/` – Database abstractions and PostgreSQL implementation. Includes SQL access layers for guilds, command logs, processed events, and module repositories; `migrations/` hosts schema evolution scripts.
- `internal/module/` – Production modules built on the Mio API. Each subdirectory is a feature area:
  - `administration` handles owner utilities such as DM forwarding and runtime command toggling.
  - `moderation` provides bans, warns, filtering, lockdowns, timeouts, and autorole hooks, combining command handlers with guild and member persistence.
  - `utility` exposes diagnostics (ping, stats), profile lookups, server metadata, color utilities, and interactive help flows with message components.
  - `search` integrates external services (YouTube, OpenWeather) and demonstrates message-component handlers for paginated results.
  - `fishing` manages collection gameplay with rarity-weighted rolls persisted via the module’s service layer.
  - `customrole`, `fun`, and `testing` showcase custom role management, recreational commands, and slash-command scaffolding respectively.
- `internal/structs/` – Shared domain models (guild settings, command log entries) and configuration loading that merges JSON defaults with environment overrides.
- `internal/utils/` – Extra helpers used by modules (e.g., unit conversion utilities for embeds).
- `pkg/mio/` – Core reusable framework for Discord bots (see “Mio API Deep Dive”).
- `pkg/utils/` – Cross-cutting utilities (configuration loader, discord helpers, embed/message builders) used by both the framework and modules.
- `assets/`, `build/`, `Dockerfile`, `docker-compose.yml`, `entrypoint.sh` – Operational assets for containerised deployment, static resources, and local development.

## Mio API Deep Dive (`pkg/mio`)

The `pkg/mio` package family abstracts DiscordGo and exposes a composable bot runtime:

### `pkg/mio/bot`

- **Bot construction**: `bot.NewBotBuilder` assembles a `Bot`, wiring Discord sessions, module manager, cooldown and callback managers, and a shared `EventBus`. Builders allow injecting custom loggers or sessions and opting into default gateway handlers (ready, guild join/leave, member chunk) defined in `bot.go`.
- **Event handling**: `EventHandler.Listen` multiplexes Discord messages and interactions to modules. It publishes internal events (e.g., `MessageProcessed`) onto the `EventBus`, enabling analytics or metrics subscribers.
- **Module system**: `Module` and `ModuleBase` implement command dispatch mechanics. During `Hook`, modules call registration helpers (`RegisterCommands`, `RegisterApplicationCommands`, `RegisterPassives`, etc.), which the base class stores and later uses to process incoming Discord events. Cooldown scopes, permission checks, and panic recovery are handled centrally in `ModuleBase`.
- **Builders**: `ModuleCommandBuilder`, `ModuleApplicationCommandBuilder`, `ModulePassiveBuilder`, and friends simplify assembling commands with fluent configuration (cooldowns, DM permissions, slash command metadata, etc.).
- **Event definitions**: `events.go` enumerates lifecycle events (command ran, panicked, etc.), allowing modules or analytics consumers to subscribe via the event bus.
- **ModuleManager**: Registers and stores modules, offers lookup helpers (`FindCommand`, `FindPassive`, …), and logs registration outcomes.

The moderation, utility, and other modules in `internal/module` leverage this API extensively—see `internal/module/moderation/moderation.go` for command registration and event hooks, or `internal/module/utility/utility.go` for a comprehensive example spanning classic commands, slash commands, and component handlers.

### `pkg/mio/discord`

- Wraps DiscordGo sessions (`Discord`, `SessionWrapper`) and exposes channel-based message/interaction streams consumed by `EventHandler`.
- Provides rich wrapper types (`DiscordMessage`, `DiscordInteraction`, `DiscordApplicationCommand`, etc.) that standardise replies, embed/file responses, permission checks, and argument parsing. Message wrappers also surface utility methods (`Args`, `CallbackKey`, `BotHasPermissions`).
- Manages shard creation, intents, and gateway listeners (`createSessions`, `Run`, `Close`), and re-exports DiscordGo request helpers on the session interface for modules that need raw access.

### `pkg/mio/utils`

- **CallbackManager**: Supports ad-hoc message-based workflows (e.g., waiting on follow-up input) keyed by `channel:user` identifiers.
- **CooldownManager**: Centralises rate limiting keyed by custom scopes (user, channel, guild). Modules rely on this to enforce fair usage across commands.

### `pkg/mio/event_bus.go`

- A lightweight reflection-based pub/sub hub. Modules or the analytics layer register handlers and receive typed events emitted by the bot runtime.

### `pkg/mio/logger.go`

- Defines the Mio logging interface with JSON field support. Modules typically request child loggers via `Logger.Named`, ensuring consistent structured logging across the bot and exposing adapters for zap in `internal/meido/logger.go`.

## Application Layer Details

- **Module life-cycle**: Every module embeds `bot.ModuleBase`, names itself, and narrows allowed message types/DM behaviour. During `Hook` it uses Mio builders to register commands. For example, `internal/module/administration/administration.go` registers a DM-forwarding passive and owner-only management commands, while `internal/module/search/search.go` registers slash commands and message component callbacks to paginate image search results.
- **Database-backed services**: Modules that need persistence inject repositories through constructors. `internal/module/fishing` wraps an `IAquariumDB` to manage player collections, while `internal/module/moderation` composes warn/filter repositories to enforce guild policies. `internal/database` offers shared primitives (`IGuildDB`, `ICommandLogDB`, `IProcessedEventsDB`) that modules can extend with their own interfaces.
- **Event analytics**: `internal/meido/logs.go` listens on the Mio event bus to persist command and interaction counts via the processed-events table, demonstrating how to consume `pkg/mio` events outside of modules.
- **Configuration**: `internal/structs.LoadConfig` loads `cmd/meido/config.json` (see `cmd/meido/config.json` for shape) and overlays environment variables, populating tokens, shard counts, owner IDs, API keys, and module exclusions.

## Data and Infrastructure

- **Database**: The bot expects a PostgreSQL database configured via `connection_string`. `internal/database/migrations/` contains schema definitions for guilds, command logs, processed events, warns, filters, fishing data, and custom roles.
- **Docker & scripts**: Docker artefacts enable containerised deployment. `docker-compose.yml` combines the bot and database for local runs. `entrypoint.sh` handles bootstrapping inside containers.
- **Assets**: Static art under `assets/` is used in marketing materials and README visuals.

## Working With the Project

- **Running locally**: Populate `cmd/meido/config.json` (or environment variables) with Discord and database credentials, run migrations, then execute `go run ./cmd/meido`. The process will connect to Discord, register slash commands, and begin handling events.
- **Extending via Mio**:
  1. Create a new package under `internal/module/<feature>` embedding `bot.ModuleBase`.
  2. In `Hook`, register commands, passives, slash commands, and component handlers using Mio builders.
  3. Inject dependencies (DB interfaces, services) through the module constructor.
  4. Register the module inside `internal/meido/meido.go`’s `registerModules` list.
- **Using Mio externally**: Because `pkg/mio` is decoupled from `internal`, it can power other bots. Refer to modules in `internal/module` for idiomatic usage patterns and to tests in `pkg/mio/bot` for builder and event bus expectations.

## Testing and Quality

- The repository includes unit tests for framework components (`pkg/mio/bot`, `pkg/mio/utils`, `pkg/utils/builders`, Discord wrappers) to validate command execution, event routing, and helper utilities.
- Module-level behaviour is validated through targeted tests (e.g., fishing services, message builders). Use `go test ./...` to execute the suite.
- Logging is structured and namespaced to aid observability; consider extending `internal/meido/logs.go` if additional metrics are required.

## Further Reading

- Start with `pkg/mio/bot` and `pkg/mio/discord` to understand the core runtime.
- Explore `internal/module/utility/` for a comprehensive showcase of commands, slash commands, and component handling.
- Review `internal/database` and `internal/module/moderation` together to see how persistence and modules interact.
- Consult `README.md` for feature highlights and links to the live bot and support resources.

