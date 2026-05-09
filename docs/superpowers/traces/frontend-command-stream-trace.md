# Frontend Command And Stream Trace

## Trace Scope

- `ui/commands/ssh.js`
- `ui/commands/telnet.js`
- `ui/control/telnet.js`
- `ui/stream/reader.js`
- `ui/stream/sender.js`
- `ui/stream/streams.js`

## Notes

- Trace connect, resize, send, read, and close behavior across the command and stream boundary.
- Keep this file as the narrow reference for Stage 4 frontend characterization.

## Command Shape

- SSH connect commands are built by `ui/commands/ssh.js` from the prompt user,
  parsed host/port, and authentication type. The initial payload is the encoded
  user string, encoded address, and one authentication byte.
- SSH terminal resize commands use stream marker `0x01` and encode terminal
  rows followed by columns as two unsigned 16-bit big-endian values.
- Telnet connect commands are built by `ui/commands/telnet.js` from the parsed
  host/port address only.
- Telnet currently has no terminal resize API in its command handler. NAWS
  resize negotiation is implemented by `ui/control/telnet.js`.
- Telnet `Control.resize()` sends `WILL NAWS`; after the server replies
  `DO NAWS`, it writes columns followed by rows as unsigned 16-bit big-endian
  values inside one NAWS subnegotiation frame.
- If the server initiates `DO NAWS` before a client resize request, Telnet writes
  `WILL NAWS` and the NAWS subnegotiation in one combined frame.

## Stream Behavior

- `ui/stream/reader.js` preserves unread bytes after partial buffer exports and
  rejects pending and later reads with the explicit close reason.
- `ui/stream/sender.js` batches outgoing bytes until the segment size, request
  count, timer, or close path forces a flush.
- `ui/stream/streams.js` owns stream registration and teardown; Stage 4 keeps it
  intact and removes only the assertion-free placeholder test file.
