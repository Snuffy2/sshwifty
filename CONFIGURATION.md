# Sshwifty Configuration

Sshwifty can be configured through either a JSON configuration file or
environment variables. By default, the configuration loader tries default file
paths first, then falls back to environment variables.

Use `SSHWIFTY_CONFIG` to specify a configuration file:

```sh
SSHWIFTY_CONFIG=./sshwifty.conf.json ./sshwifty
```

This tells Sshwifty to load configuration from `./sshwifty.conf.json`.

## Configuration File

`sshwifty.conf.example.json` is an example of a valid configuration file. Use it
as a starting point for your own configuration.

```jsonc
{
  // HTTP Host. Keep it empty to accept request from all hosts, otherwise, only
  // specified host is allowed to access
  "HostName": "localhost",

  // Web interface access password. Set to empty to allow public access to the
  // web interface (bypass the Authenticate page)
  "SharedKey": "WEB_ACCESS_PASSWORD",

  // Remote dial timeout. This limits how long of time the backend can spend
  // to connect to a remote host. The max timeout will be determined by
  // server configuration (ReadTimeout).
  // (In Seconds)
  "DialTimeout": 10,

  // Socks5 proxy. When set, Sshwifty backend will try to connect remote through
  // the given proxy
  "Socks5": "localhost:1080",

  // Username of the Socks5 server. Please set when needed
  "Socks5User": "",

  // Password of the Socks5 server. Please set when needed
  "Socks5Password": "",

  // Server-side hooks, allowing operators to launch external processes on the
  // server side to influence server behavior
  //
  // The operation of a Hook must be completed within the time limit defined
  // by `HookTimeout` set below. Otherwise it will be terminated, and results
  // a failure for the execution
  //
  // To determine how much time is still left for the execution, a Hook can
  // fetch the deadline information from the `SSHWIFTY_HOOK_DEADLINE`
  // environment variable which is a RFC3339 formatted date string indicating
  // after what time the termination will occur
  //
  // Warning: the process will be launched within the same context and system
  // permission which Sshwifty is running under, thus is it crucial that the
  // Hook process is designed and operated in a secure manner, otherwise
  // SECURITY VULNERABILITY (commandline injection, for example) maybe created
  // as result
  //
  // Warning: all inputs passed by Sshwifty to the hook process must be
  // considered unsanitized, and must be sanitized by each hook themselves
  "Hooks": {
    // before_connecting is called before Sshwifty starts to connect to a remote
    // endpoint. If any of the Hook process exited with a non-zero return code,
    // the connection request is aborted
    //
    // This Hook offers two parameters:
    // - SSHWIFTY_HOOK_REMOTE_TYPE: Type of the connection (i.e. SSH or Telnet)
    // - SSHWIFTY_HOOK_REMOTE_ADDRESS: Address of the remote host
    "before_connecting": [
      // Following example command launches a `/bin/sh` to execute a for loop
      // that prints to Stdout as well as to Stderr
      //
      // Prints to Stdout will be sent to the client side visible to the user,
      // and prints to Stderr will be captured as server side logs and it is
      // invisible to the user (as server logs usually are)
      //
      // The command must be specified in Json array format. Each array element
      // is mapped to a command fragment separated by space. For example:
      // ["command", "-i", "Hello World"] will be mapped to `command -i "Hello
      // World"` before it is executed
      [
        "/bin/sh",
        "-c",
        "for n in $(seq 1 5); do sleep 1 && echo Stdout $SSHWIFTY_HOOK_REMOTE_TYPE $n && echo Stderr $SSHWIFTY_HOOK_REMOTE_TYPE $n 1>&2; done",
      ],
      // You can add multiple hooks, they're executed in sequence even when the
      // previous one fails
      ["/bin/sh", "-c", "/etc/sshwifty/before_connecting.sh"],
      ["/bin/another-command", "...", "..."],
    ],
  },

  // The maximum execution time of each hook, in seconds. If this timeout is
  // exceeded, the hook will be terminated, and thus cause a failure
  "HookTimeout": 30,

  // Sshwifty HTTP server, you can set multiple ones to serve on different
  // ports
  "Servers": [
    {
      // Which local network interface this server will be listening
      "ListenInterface": "0.0.0.0",

      // Which local network port this server will be listening
      "ListenPort": 8182,

      // Timeout of initial request. HTTP handshake must be finished within
      // this time
      // (In Seconds)
      "InitialTimeout": 10,

      // How long do the connection can stay in idle before the backend server
      // disconnects the client
      // (In Seconds)
      "ReadTimeout": 120,

      // How long the server will wait until the client connection is ready to
      // receive new data. If this timeout is exceeded, the connection will be
      // closed.
      // (In Seconds)
      "WriteTimeout": 120,

      // The interval between internal echo requests
      // (In Seconds)
      "HeartbeatTimeout": 10,

      // Forced delay between each request
      // (In Milliseconds)
      "ReadDelay": 10,

      // Forced delay between each write
      // (In Milliseconds)
      "WriteDelay": 10,

      // Path to TLS certificate file. Set empty to use HTTP
      "TLSCertificateFile": "",

      // Path to TLS certificate key file. Set empty to use HTTP
      "TLSCertificateKeyFile": "",

      // Display a short text message on the Home page. Link is supported
      // through `[Title text](https://link.example.com)` format
      "ServerMessage": "",
    },
    {
      "ListenInterface": "0.0.0.0",
      "ListenPort": 8183,
      "InitialTimeout": 3,
    },
  ],

  // Remote Presets, the operator can define presets for users so the user
  // won't have to manually fill-in all the form fields
  //
  // Presets will be displayed in the "Known remotes" tab on the Connector
  // window
  //
  // Notice: You can use the same JSON value for `SSHWIFTY_PRESETS` if you are
  //         configuring your Sshwifty through environment variables.
  //
  // Warning: Presets Data will be sent to user client WITHOUT any protection.
  //          DO NOT add any secret information into Preset.
  "Presets": [
    {
      // Title of the preset
      "Title": "SDF.org Unix Shell",

      // Preset Types, i.e. Telnet, SSH, and Mosh
      "Type": "SSH",

      // Target address and port
      "Host": "sdf.org:22",

      // Define the tab and background color of the console in RGB hex format
      // for better visual identification
      //
      // For example: 110000 will give you a dark red background, 001100 is
      // dark green and 000011 is dark blue
      //
      // The color must not be too bright, as it will make the foreground text
      // hard to read
      "TabColor": "112233",

      // Form fields and values, you have to manually validate the correctness
      // of the field value
      //
      // Defining a Meta field will prevent user from changing it on their
      // Connector Wizard. If you want to allow users to use their own settings,
      // leave the field unset
      //
      // Values in Meta are scheme enabled, and supports following scheme
      // prefixes:
      // - "literal://": Text literal (Default)
      //                 Example: literal://Data value
      //                          (The final value will be "Data value")
      //                 Example: literal://file:///tmp/afile
      //                          (The final value will be "file:///tmp/afile")
      // - "file://": Load Meta value from given file.
      //              Example: file:///home/user/.ssh/private_key
      //                       (The file path is /home/user/.ssh/private_key)
      // - "environment://": Load Meta value from an Environment Variable.
      //                    Example: environment://PRIVATE_KEY_DATA
      //                    (The name of the target environment variable is
      //                    PRIVATE_KEY_DATA)
      //
      // All data in Meta is loaded during start up, and will not be updated
      // even the source already been modified.
      "Meta": {
        // Data for predefined User field
        "User": "pre-defined-username",

        // Data for predefined Encoding field. Valid data is those displayed on
        // the page
        "Encoding": "pre-defined-encoding",

        // Data for predefined Mosh Server field. Defaults to "mosh-server"
        "Mosh Server": "mosh-server",

        // Data for predefined Password field
        "Password": "pre-defined-password",

        // Data for predefined Private Key field, should contains the content
        // of a Key file
        "Private Key": "file:///home/user/.ssh/private_key",

        // Data for predefined Authentication field. Valid values is what
        // displayed on the page (Password, Private Key, None)
        "Authentication": "Password",

        // Data for server public key fingerprint. You can acquire the value of
        // the fingerprint by manually connect to a new SSH host with Sshwifty,
        // the fingerprint will be displayed on the Fingerprint confirmation
        // page.
        "Fingerprint": "SHA256:bgO....",
      },
    },
    {
      "Title": "Endpoint Telnet",
      "Type": "Telnet",
      "Host": "telnet.example.com:23",
      "Meta": {
        // Data for predefined Encoding field. Valid data is those displayed on
        // the page
        "Encoding": "utf-8",
      },
    },
    {
      "Title": "Example Mosh",
      "Type": "Mosh",
      "Host": "ssh.example.com:22",
      "Meta": {
        "User": "guest",
        "Authentication": "Password",
        "Encoding": "utf-8",
        "Mosh Server": "mosh-server",
      },
    },
  ],

  // Allow the Preset Remotes only, and refuse to connect to any other remote
  // host
  //
  // NOTICE: You can only configure OnlyAllowPresetRemotes through a config
  //         file. This option is not supported when you are configuring with
  //         environment variables
  "OnlyAllowPresetRemotes": false,
}
```

## Environment Variables

Valid environment variables are:

```text
SSHWIFTY_HOSTNAME
SSHWIFTY_SHAREDKEY
SSHWIFTY_DIALTIMEOUT
SSHWIFTY_SOCKS5
SSHWIFTY_SOCKS5_USER
SSHWIFTY_SOCKS5_PASSWORD
SSHWIFTY_HOOK_BEFORE_CONNECTING
SSHWIFTY_HOOKTIMEOUT
SSHWIFTY_LISTENPORT
SSHWIFTY_INITIALTIMEOUT
SSHWIFTY_READTIMEOUT
SSHWIFTY_WRITETIMEOUT
SSHWIFTY_HEARTBEATTIMEOUT
SSHWIFTY_READDELAY
SSHWIFTY_WRITEDELAY
SSHWIFTY_LISTENINTERFACE
SSHWIFTY_TLSCERTIFICATEFILE
SSHWIFTY_TLSCERTIFICATEKEYFILE
SSHWIFTY_SERVERMESSAGE
SSHWIFTY_PRESETS
SSHWIFTY_ONLYALLOWPRESETREMOTES
```

These options correspond to their counterparts in the configuration file.

`SSHWIFTY_PRESETS` must contain valid JSON-encoded preset data. Its format is
shown in [preset.example.json](preset.example.json) and can be loaded with:

```sh
SSHWIFTY_PRESETS="$(cat preset.example.json)" ./sshwifty
```

You can also set `SSHWIFTY_PRESETS` directly as a string. In that case you may
need to escape the JSON characters. One option is:

```sh
jq -c . preset.example.json | jq -Rs
```

When using environment variables, only one Sshwifty HTTP server is allowed. Use
the configuration file if you need to serve on multiple ports.

Invalid values in these environment variables are silently reset to defaults
during configuration parsing:

```text
SSHWIFTY_DIALTIMEOUT
SSHWIFTY_INITIALTIMEOUT
SSHWIFTY_READTIMEOUT
SSHWIFTY_WRITETIMEOUT
SSHWIFTY_HEARTBEATTIMEOUT
SSHWIFTY_READDELAY
SSHWIFTY_WRITEDELAY
```

Verify these values before starting the instance.
