package e2e

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"math/rand"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	"gotest.tools/v3/assert"
)

var api *httpexpect.Expect
var baseURL string

func TestMain(m *testing.M) {
	baseURL = "http://localhost:8080"

	api = httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  baseURL,
		Reporter: httpexpect.NewPanicReporter(),
		Client:   &http.Client{Timeout: 30 * time.Second},
	})

	waitAPI()

	code := m.Run()
	os.Exit(code)
}

func waitAPI() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-time.After(30 * time.Second):
			log.Fatalf("Timeout: API not ready at %s", baseURL)
		case <-ticker.C:
			resp := api.GET("/ping").Expect().Raw()
			if resp.StatusCode == 200 {
				return
			}
		}
	}
}

type ClientUpsert struct {
	ClientID string `json:"client_id"`
	Login    string `json:"login"`
	Age      int    `json:"age"`
	Location string `json:"location"`
	Gender   string `json:"gender"`
}

func TestClientLifecycle(t *testing.T) {
	clientID := uuid.New().String()

	t.Run("Create Client", func(t *testing.T) {
		clients := []ClientUpsert{
			{
				ClientID: clientID,
				Login:    "created_client",
				Age:      21,
				Location: "Moscow",
				Gender:   "MALE",
			},
		}

		response := api.POST("/clients/bulk").
			WithJSON(clients).
			Expect().
			Status(http.StatusCreated).
			JSON().Array()

		assert.Equal(t, 1, len(response.Raw()), "Ожидался ответ с одним клиентом")
		response.Value(0).Object().Value("client_id").IsEqual(clientID)
	})

	time.Sleep(1 * time.Second)
	t.Run("Get Client By ID", func(t *testing.T) {
		client := api.GET("/clients/{clientId}", clientID).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		client.Value("client_id").IsEqual(clientID)
		client.Value("login").IsEqual("created_client")
		client.Value("age").IsEqual(21)
		client.Value("location").IsEqual("Moscow")
		client.Value("gender").IsEqual("MALE")
	})

	t.Run("Update Client", func(t *testing.T) {
		clients := []ClientUpsert{
			{
				ClientID: clientID,
				Login:    "updated_client",
				Age:      30,
				Location: "Vladimir",
				Gender:   "MALE",
			},
		}

		log.Println(clients)
		// Perform update
		response := api.POST("/clients/bulk").
			WithJSON(clients).
			Expect().
			Status(http.StatusCreated).
			JSON().Array()

		// Verify the update response
		assert.Equal(t, 1, len(response.Raw()), "Ожидался ответ с одним клиентом")
		response.Value(0).Object().Value("client_id").IsEqual(clientID)
		log.Println(response.Raw())

		time.Sleep(1 * time.Second)
		// Verify the updated data
		client := api.GET("/clients/{clientId}", clientID).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		log.Println(client.Raw())
		client.ContainsKey("client_id").ContainsKey("login")
		assert.Equal(t, clientID, client.Value("client_id").Raw())
		assert.Equal(t, "updated_client", client.Value("login").Raw())
		assert.Equal(t, 30, client.Value("age").Raw())
		assert.Equal(t, "Vladimir", client.Value("location").Raw())
		assert.Equal(t, "MALE", client.Value("gender").Raw())
	})

	t.Run("Get Non-Existing Client", func(t *testing.T) {
		api.GET("/clients/{clientId}", uuid.New().String()).
			Expect().
			Status(http.StatusNotFound)
	})

	t.Run("Invalid Client Data", func(t *testing.T) {
		invalidClients := []map[string]interface{}{
			{
				"client_id": "",
				"login":     "invalid_client",
				"age":       -1,
				"location":  "Unknown",
				"gender":    "INVALID",
			},
			{
				"client_id": "not-a-uuid",
				"login":     "",
				"age":       150,
				"gender":    "MALE",
			},
		}

		response := api.POST("/clients/bulk").
			WithJSON(invalidClients).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object()

		response.ContainsKey("message")
	})

	t.Run("Batch Create Clients", func(t *testing.T) {
		clients := []ClientUpsert{
			{
				ClientID: uuid.New().String(),
				Login:    "batch_client_1",
				Age:      25,
				Location: "Moscow",
				Gender:   "MALE",
			},
			{
				ClientID: uuid.New().String(),
				Login:    "batch_client_2",
				Age:      30,
				Location: "Saint Petersburg",
				Gender:   "FEMALE",
			},
		}

		response := api.POST("/clients/bulk").
			WithJSON(clients).
			Expect().
			Status(http.StatusCreated).
			JSON().Array()

		assert.Equal(t, 2, len(response.Raw()), "Ожидался ответ с двумя клиентами")

		// Проверяем каждого созданного клиента
		for _, client := range clients {
			time.Sleep(time.Second)
			resp := api.GET("/clients/{clientId}", client.ClientID).
				Expect().
				Status(http.StatusOK).
				JSON().Object()

			resp.Value("client_id").IsEqual(client.ClientID)
			resp.Value("login").IsEqual(client.Login)
			resp.Value("age").IsEqual(client.Age)
			resp.Value("location").IsEqual(client.Location)
			resp.Value("gender").IsEqual(client.Gender)
		}
	})

	t.Run("Empty Request Body", func(t *testing.T) {
		resp := api.POST("/clients/bulk").
			WithJSON([]ClientUpsert{}).
			Expect().
			Status(http.StatusCreated)
		resp.Body().Contains("[]")
	})

	t.Run("Large Batch Update", func(t *testing.T) {
		// Создаем большой пакет клиентов
		var clients []ClientUpsert
		for i := 0; i < 10; i++ {
			clients = append(clients, ClientUpsert{
				ClientID: uuid.New().String(),
				Login:    fmt.Sprintf("bulk_client_%d", i),
				Age:      20 + i,
				Location: "Moscow",
				Gender:   "MALE",
			})
		}

		// Отправляем большой пакет
		response := api.POST("/clients/bulk").
			WithJSON(clients).
			Expect().
			Status(http.StatusCreated).
			JSON().Array()

		assert.Equal(t, len(clients), len(response.Raw()), "Количество созданных клиентов не совпадает с ожидаемым")

		// Проверяем несколько случайных клиентов
		time.Sleep(time.Second)
		for i := 0; i < 3; i++ {
			randomIndex := rand.Intn(len(clients))
			client := clients[randomIndex]

			resp := api.GET("/clients/{clientId}", client.ClientID).
				Expect().
				Status(http.StatusOK).
				JSON().Object()

			resp.Value("login").IsEqual(client.Login)
			resp.Value("age").IsEqual(client.Age)
		}
	})
}

// Добавляем вспомогательную функцию для создания тестовых данных
func createTestClient(api *httpexpect.Expect) (string, error) {
	clientID := uuid.New().String()
	client := []ClientUpsert{
		{
			ClientID: clientID,
			Login:    "test_client",
			Age:      25,
			Location: "Test City",
			Gender:   "MALE",
		},
	}

	api.POST("/clients/bulk").
		WithJSON(client).
		Expect().
		Status(http.StatusCreated)

	time.Sleep(time.Second)
	return clientID, nil
}
