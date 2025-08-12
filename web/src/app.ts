import { Device } from "@twilio/voice-sdk";
import { BASE_URL } from "./config";
import type { Call } from "@twilio/voice-sdk";

// TODO: get from log in
const IDENTITY = "user";

let device: Device;

document.addEventListener("DOMContentLoaded", () => {
    const connectBtn = document.getElementById(
        "connect-btn"
    ) as HTMLButtonElement;

    const callBtn = document.getElementById("call-btn") as HTMLButtonElement;

    let call: Call | null = null;

    connectBtn.onclick = async () => {
        connectBtn.style.display = "none";

        try {
            const url: string = BASE_URL + "/token?identity=" + IDENTITY;
            console.log("Fetching token from:", url);

            const resp = await fetch(url, {
                headers: {
                    "ngrok-skip-browser-warning": "true",
                },
            });

            const { token } = await resp.json();
            console.log("Received token:", token);

            device = new Device(token);

            device.on("registered", () => {
                console.log("Device registered");
                callBtn.disabled = false;
            });

            device.on("error", (err) => {
                console.error("Twilio Device Error:", err);
            });

            device.register();
        } catch (err) {
            console.error("Failed to initialize Twilio Device:", err);
        }
    };

    callBtn.onclick = async () => {
        const toInput = document.getElementById("to-input") as HTMLInputElement;
        const destNumber: string = toInput.value.trim();

        if (destNumber === "") {
            alert("Please enter a destination number.");
            return;
        }

        if (!device) {
            alert("Device is not initialized. Please connect first.");
            return;
        }

        if (call !== null) {
            console.log("Already in a call, hanging up first.");
            call.disconnect();
            call = null;
        }

        call = await device.connect({ params: { To: destNumber } });
    };
});
