Shrimple setup for making outbound calls through Twilio via their browser SDK. Didn't end up developing it further because we shifted to Telnyx due to costs.

After running the server (`make` inside the srvr/ directory), make it publicly accessible via ngrok (`ngrok http 8065`) and set that URL in your TwiML app (to serve as a webhook), in addition to web/config.ts (for the JWT API). Finally, serve the webpage w/ `pnpm run serve`.
