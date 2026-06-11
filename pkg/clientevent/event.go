package clientevent

// Топик по умолчанию для событий регистрации клиентов.
const TopicDefault = "client.registration.v1"

// Registered — тип события успешной регистрации клиента.
const Registered = "client.registered"

// RegisteredEvent публикуется client-registration-service.
// В Kafka передаётся только password_hash, никогда plain password.
type RegisteredEvent struct {
	Event        string `json:"event"`
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
	FullName     string `json:"full_name"`
	Phone        string `json:"phone"`
	VehicleID    string `json:"vehicle_id,omitempty"`
}
