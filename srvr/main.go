package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/twilio/twilio-go/client/jwt"
	"github.com/twilio/twilio-go/twiml"
)

func generateToken(identity string) (string, error) {
	// Identity is generally a unique identifier for the user, e.g a username

	voiceGrant := &jwt.VoiceGrant{
		Incoming: jwt.Incoming{Allow: false},
		Outgoing: jwt.Outgoing{ApplicationSid: os.Getenv("TWILIO_TWIML_APP_SID")},
	}

	token := jwt.CreateAccessToken(jwt.AccessTokenParams{
		AccountSid:    os.Getenv("TWILIO_ACCOUNT_SID"),
		SigningKeySid: os.Getenv("TWILIO_API_KEY_SID"),
		Secret:        os.Getenv("TWILIO_API_KEY_SECRET"),
		Identity:      identity,
	})

	token.AddGrant(voiceGrant)

	return token.ToJwt()
}

// Client-side will hit this handler and get returned a JWT AKA voice grant
// It'll then use that JWT to make calls via Twilio
func tokenHandler(w http.ResponseWriter, r *http.Request) {
	identity := r.URL.Query().Get("identity")
	if identity == "" {
		http.Error(w, "identity required", http.StatusBadRequest)
		return
	}

	token, err := generateToken(identity)
	if err != nil {
		log.Print("Error generating token: ", err)
		http.Error(w, "error generating token", http.StatusInternalServerError)
		return
	}

	// log.Printf("Generated token `%s` for identity: %s", token, identity)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// Twilio will hit this handler when the client-side app attempts to make a call
// It will return a TwiML response that instructs Twilio to dial the number
func voiceHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	to := r.FormValue("To")
	if to == "" {
		http.Error(w, "Missing 'To' parameter", http.StatusBadRequest)
		return
	}

	// TODO: load this from DB given the user identity string
	number := os.Getenv("TWILIO_CALLER_NUMBER")

	vd := &twiml.VoiceDial{
		Number:   to,
		CallerId: number,
	}

	twimlResp, err := twiml.Voice([]twiml.Element{vd})
	if err != nil {
		http.Error(w, "Failed to render TwiML", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(twimlResp))
}

func main() {
	godotenv.Load()

	if os.Getenv("TWILIO_ACCOUNT_SID") == "" ||
		os.Getenv("TWILIO_API_KEY_SID") == "" ||
		os.Getenv("TWILIO_API_KEY_SECRET") == "" ||
		os.Getenv("TWILIO_CALLER_NUMBER") == "" ||
		os.Getenv("TWILIO_TWIML_APP_SID") == "" {
		log.Println("Please set the appropriate environment variables")
		log.Fatal(
			"\n\tTWILIO_ACCOUNT_SID: ", os.Getenv("TWILIO_ACCOUNT_SID"), "\n",
			"\tTWILIO_API_KEY_SID: ", os.Getenv("TWILIO_API_KEY_SID"), "\n",
			"\tTWILIO_API_KEY_SECRET: ", os.Getenv("TWILIO_API_KEY_SECRET"), "\n",
			"\tTWILIO_CALLER_NUMBER: ", os.Getenv("TWILIO_CALLER_NUMBER"), "\n",
			"\tTWILIO_TWIML_APP_SID: ", os.Getenv("TWILIO_TWIML_APP_SID"), "\n",
		)
	}

	http.HandleFunc("/token", tokenHandler)
	http.HandleFunc("/voice", voiceHandler)

	corsHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Access-Control-Allow-Origin", "*")
			w.Header().Add("Access-Control-Allow-Headers", "ngrok-skip-browser-warning, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	fmt.Println("Server running on http://localhost:8065")
	log.Fatal(http.ListenAndServe(":8065", corsHandler(http.DefaultServeMux)))
}
