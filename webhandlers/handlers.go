package webhandlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
	"worker-bot/models"

	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"
	"github.com/jmoiron/sqlx"
	"github.com/k0kubun/pp"
	"golang.org/x/exp/rand"
	"google.golang.org/api/option"
)

type HandlerV1 struct {
	db *sqlx.DB
}

func NewHandlerV1(db *sqlx.DB) *HandlerV1 {
	return &HandlerV1{
		db: db,
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type Response map[string]interface{}

// @Summary     Get Questions
// @Description This API generates 10 questions with Gemini AI providing their answers too.
// @Tags  	    Question
// @Accept      json
// @Produce     json
// @Param       difficulty path string true "Difficulty Level" Enums(easy, medium, hard)
// @Success     200 {object} ErrorResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /questions/{difficulty} [get]
func (h *HandlerV1) TestGenHandler(c *gin.Context) {
	difficulty := c.Param("difficulty")
	client, err := genai.NewClient(context.Background(), option.WithAPIKey("AIzaSyCEBU4MIMl2bMzBDR9ZPDjW-8k0JBVZEMM"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error while creating a new client",
		})
		return
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")

	model.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockOnlyHigh,
		},
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: genai.HarmBlockOnlyHigh,
		},
	}

	model.GenerationConfig = genai.GenerationConfig{
		ResponseMIMEType: "application/json",
	}

	prompt := fmt.Sprintf(`{
        "tests": [
            {
                "question": "Which of the following is NOT a category within the broad field of Ecology?",
                "variants": [
                    {
                        "A": "Biomes"
                    },
                    {
                        "B": "Ecosystems"
                    },
                    {
                        "C": "Biodiversity"
                    },
                    {
                        "D": "Astrophysics"
                    }
                ]
            }
        ],
        "answers": {
            "1": "A",
            "2": "B",
            "3": "C",
            "4": "D",
            "5": "A",
            "6": "B",
            "7": "C",
            "8": "D",
            "9": "A",
            "10": "B"
        }
    }
    GENERATE ME 10 RANDOM ECOLOGY TESTS APPLYING THIS FORMAT ABOVE. QUESTION NUMBERS ARE DYNAMIC. I WILL GIVE YOU DIFFICULTY OF QUESTIONS. IT MAY BE EASY, MEDIUM or HARD. So DIFFICULTY LEVEL IS: %s`, difficulty)

	resp, err := model.GenerateContent(context.Background(), genai.Text(prompt))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error while generating tests",
		})
		return
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		c.JSON(http.StatusNoContent, gin.H{
			"error": "No content from Gemini",
		})
		return
	}

	temp, err := json.Marshal(resp.Candidates[0].Content.Parts[0])
	if err != nil {
		log.Fatal(err)
	}

	unescapedData, err := strconv.Unquote(string(temp))
	if err != nil {
		log.Fatal(err)
	}

	var questions map[string]interface{}
	err = json.Unmarshal([]byte(unescapedData), &questions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error while unmarshaling answer to map" + err.Error(),
		})
		return
	}

	// Extract questions and variants for translation
	tests := questions["tests"].([]interface{})
	var texts []string
	for _, test := range tests {
		question := test.(map[string]interface{})["question"].(string)
		texts = append(texts, question)
		for _, variant := range test.(map[string]interface{})["variants"].([]interface{}) {
			for _, variantValue := range variant.(map[string]interface{}) {
				texts = append(texts, variantValue.(string))
			}
		}
	}

	translatedTexts, err := translateTexts(texts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error while translating questions" + err.Error(),
		})
		return
	}

	textIndex := 0
	for _, test := range tests {
		test.(map[string]interface{})["question"] = translatedTexts[textIndex]
		textIndex++
		for _, variant := range test.(map[string]interface{})["variants"].([]interface{}) {
			for variantKey := range variant.(map[string]interface{}) {
				variant.(map[string]interface{})[variantKey] = translatedTexts[textIndex]
				textIndex++
			}
		}
	}

	c.JSON(http.StatusOK, questions)
}
func translateTexts(texts []string) ([]string, error) {
	url := "https://websocket.tahrirchi.uz/translate"
	payload := map[string]interface{}{
		"text": map[string]interface{}{
			"texts": texts,
		},
		"source_lang": "eng_Latn",
		"target_lang": "uzn_Latn",
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "389eebc7-4e87-4c59-b0c0-d1a1f1c0aacc")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("Translation service response: %s", string(body))

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	sentences, ok := response["sentences"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	var translatedTexts []string
	for _, sentence := range sentences {
		translatedText, ok := sentence.(map[string]interface{})["translated"].(string)
		if !ok {
			return nil, fmt.Errorf("unexpected sentence format")
		}
		translatedTexts = append(translatedTexts, translatedText)
	}

	return translatedTexts, nil
}

// @Summary     Get Rankings
// @Description This API returns the ranking of users based on XP
// @Tags  	    Ranking
// @Accept      json
// @Produce     json
// @Success     200 {object} []models.RankingResponse
// @Failure     400 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /ranking [get]
func (h *HandlerV1) GetRanking(c *gin.Context) {
	var users []models.User
	err := h.db.Select(&users, "SELECT id, first_name, last_name, avatar, xp, location FROM users ORDER BY xp DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching data"})
		pp.Println(err.Error())
		return
	}

	rankings := make([]models.RankingResponse, len(users))
	for i, user := range users {
		rankings[i] = models.RankingResponse{
			ID:       user.ID,
			Rank:     i + 1,
			UserName: user.FirstName + " " + user.LastName,
			XP:       user.XP,
			Avatar:   "https://i.pravatar.cc/150?img=" + strconv.Itoa(i),
			Location: user.Location,
		}
	}

	c.JSON(http.StatusOK, gin.H{"rankings": rankings})
}

//User------------------------------

// @Summary     Create User
// @Description This API creates a new user
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        user  body models.User  true  "User Data"
// @Success      201   {object} models.User
// @Failure      400   {object} ErrorResponse
// @Failure      500   {object} ErrorResponse
// @Router       /user [post]
func (h *HandlerV1) CreateUser(c *gin.Context) {
	var user models.User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	query := `INSERT INTO users (id, first_name, last_name, avatar, birth_date, location, phone_number, xp) 
			  VALUES (:id, :first_name, :last_name, :avatar, :birth_date, :location, :phone_number, :xp)`

	_, err := h.db.NamedExec(query, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// @Summary		EarnXP
// @Description Adds XP by given data
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        user  body models.EarnXP  true  "XP Data"
// @Success      201   {object} models.Message
// @Failure      400   {object} ErrorResponse
// @Failure      500   {object} ErrorResponse
// @Router       /xp [post]
func (h *HandlerV1) EarnXP(c *gin.Context) {
	var xp models.EarnXP
	if err := c.ShouldBindJSON(&xp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		pp.Println(err.Error())
		return
	}

	var maxXP int
	switch xp.Difficulty {
	case "EASY":
		maxXP = 5
	case "MEDIUM":
		maxXP = 10
	case "HARD":
		maxXP = 15
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid difficulty level"})
		pp.Print(xp.Difficulty)
		return
	}

	totalQuestions := 10

	totalXP := (xp.CorrectCount * maxXP) / totalQuestions

	query := `UPDATE users SET xp = xp + $1 WHERE id = $2`
	_, err := h.db.Exec(query, totalXP, xp.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update XP"})
		pp.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "XP earned successfully", "earnedXP": totalXP})
}

// @Summary     Update User
// @Description This API updates user details
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        id    path int    true  "User ID"
// @Param        user  body models.User  true  "Updated User Data"
// @Success      200   {object} models.User
// @Failure      400   {object} ErrorResponse
// @Failure      404   {object} ErrorResponse
// @Failure      500   {object} ErrorResponse
// @Router       /user/{id} [put]
func (h *HandlerV1) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	query := `UPDATE users SET first_name = :first_name, last_name = :last_name, avatar = :avatar, 
			  birth_date = :birth_date, location = :location, phone_number = :phone_number, 
			  xp = :xp, updated_at = CURRENT_TIMESTAMP WHERE id = :id`

	_, err := h.db.NamedExec(query, map[string]interface{}{
		"id":           id,
		"first_name":   user.FirstName,
		"last_name":    user.LastName,
		"avatar":       user.Avatar,
		"birth_date":   user.BirthDate,
		"location":     user.Location,
		"phone_number": user.PhoneNumber,
		"xp":           user.XP,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// @Summary     Get User
// @Description This API returns user details based on the provided user ID
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        id   path int  true  "User ID"
// @Success      200  {object} models.User
// @Failure      400  {object} ErrorResponse
// @Failure      404  {object} ErrorResponse
// @Failure      500  {object} ErrorResponse
// @Router       /user/{id} [get]
func (h *HandlerV1) GetUser(c *gin.Context) {
	id := c.Param("id")
	query := `SELECT id, first_name, 
						last_name, avatar, 
						birth_date, location, 
						phone_number, xp FROM users WHERE id = $1`

	var user models.User
	err := h.db.Get(&user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			log.Printf("Error fetching user data: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user data"})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// @Summary     Delete User
// @Description This API deletes a user based on the provided user ID
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        id  path int  true  "User ID"
// @Success      204
// @Failure      400  {object} ErrorResponse
// @Failure      404  {object} ErrorResponse
// @Failure      500  {object} ErrorResponse
// @Router       /user/{id} [delete]
func (h *HandlerV1) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	query := `DELETE FROM users WHERE id = $1`

	_, err := h.db.Exec(query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting user"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary     List Users
// @Description This API returns a list of all users
// @Tags         User
// @Accept       json
// @Produce      json
// @Success      200  {array} models.User
// @Failure      500  {object} ErrorResponse
// @Router       /users [get]
func (h *HandlerV1) ListUsers(c *gin.Context) {
	query := `SELECT id, first_name, last_name, birth_date, location, phone_number, xp FROM users`
	var users []models.User
	err := h.db.Select(&users, query)
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching users"})
		return
	}

	for key := range users {
		users[key].Avatar = "https://i.pravatar.cc/150?img=" + strconv.Itoa(users[key].ID+3)
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

//Event------------------------------

// @Summary     Create Event
// @Description This API creates a new event
// @Tags         Event
// @Accept       json
// @Produce      json
// @Param        event  body models.Event  true  "Event Data"
// @Success      201    {object} models.Event
// @Failure      400    {object} ErrorResponse
// @Failure      500    {object} ErrorResponse
// @Router       /event [post]
func (h *HandlerV1) CreateEvent(c *gin.Context) {
	var event models.Event
	if err := c.BindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	query := `INSERT INTO events (id, image, name, description, total_xp, 
								start_date, end_date, resp_officer, resp_officer_image) 
			  					VALUES (:id, :image, :name, :description, :total_xp, :start_date, :end_date, :resp_officer, :resp_officer_image)`

	_, err := h.db.NamedExec(query, event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating event"})
		return
	}

	c.JSON(http.StatusCreated, event)
}

// @Summary     Update Event
// @Description This API updates event details
// @Tags         Event
// @Accept       json
// @Produce      json
// @Param        id     path string  true  "Event ID"
// @Param        event  body models.Event  true  "Updated Event Data"
// @Success      200    {object} models.Event
// @Failure      400    {object} ErrorResponse
// @Failure      404    {object} ErrorResponse
// @Failure      500    {object} ErrorResponse
// @Router       /event/{id} [put]
func (h *HandlerV1) UpdateEvent(c *gin.Context) {
	id := c.Param("id")
	var event models.Event
	if err := c.BindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	query := `UPDATE events SET image = :image, name = :name, description = :description, total_xp = :total_xp, 
			  start_date = :start_date, end_date = :end_date, resp_officer = :resp_officer, 
			  resp_officer_image = :resp_officer_image, updated_at = CURRENT_TIMESTAMP 
			  WHERE id = :id`

	_, err := h.db.NamedExec(query, map[string]interface{}{
		"id":                 id,
		"image":              event.Image,
		"name":               event.Name,
		"description":        event.Description,
		"total_xp":           event.TotalXP,
		"start_date":         event.StartDate,
		"end_date":           event.EndDate,
		"resp_officer":       event.RespOfficer,
		"resp_officer_image": event.RespOfficerImage,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating event"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// @Summary     Delete Event
// @Description This API deletes an event based on the provided event ID
// @Tags         Event
// @Accept       json
// @Produce      json
// @Param        id  path string  true  "Event ID"
// @Success      204
// @Failure      400  {object} ErrorResponse
// @Failure      404  {object} ErrorResponse
// @Failure      500  {object} ErrorResponse
// @Router       /event/{id} [delete]
func (h *HandlerV1) DeleteEvent(c *gin.Context) {
	id := c.Param("id")
	query := `DELETE FROM events WHERE id = $1`

	_, err := h.db.Exec(query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting event"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary     Get Event
// @Description This API returns event details based on the provided event ID
// @Tags         Event
// @Accept       json
// @Produce      json
// @Param        id   path string  true  "Event ID"
// @Success      200  {object} models.Event
// @Failure      400  {object} ErrorResponse
// @Failure      404  {object} ErrorResponse
// @Failure      500  {object} ErrorResponse
// @Router       /event/{id} [get]
func (h *HandlerV1) GetEvent(c *gin.Context) {
	id := c.Param("id")
	query := `SELECT id, name, description, total_xp, start_date, 
		end_date, resp_officer, resp_officer_image, created_at, updated_at 
              FROM events WHERE id = $1`

	var event models.Event
	err := h.db.Get(&event, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		} else {
			log.Printf("Error fetching event data: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching event data"})
		}
		return
	}

	c.JSON(http.StatusOK, event)
}

// @Summary     List Events
// @Description This API returns a list of all events
// @Tags         Event
// @Accept       json
// @Produce      json
// @Success      200  {array} models.Event
// @Failure      500  {object} ErrorResponse
// @Router       /events [get]
func (h *HandlerV1) ListEvents(c *gin.Context) {
	query := `SELECT id, image, name, description, 
				total_xp, start_date, end_date, 
				resp_officer, resp_officer_image, 
				created_at, updated_at, location FROM events`

	var events []models.Event
	err := h.db.Select(&events, query)
	if err != nil {
		log.Printf("Error fetching events: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching events"})
		return
	}

	for i := range events {
		events[i].RespOfficerImage = "https://i.pravatar.cc/150?img=" + strconv.Itoa(i)
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}

//History------------------------------

// @Summary     Create History
// @Description This API creates a new history record
// @Tags         History
// @Accept       json
// @Produce      json
// @Param        body body models.History true "History Data"
// @Success      201  {object} models.History
// @Failure      400  {object} ErrorResponse
// @Failure      500  {object} ErrorResponse
// @Router       /history [post]
func (h *HandlerV1) CreateHistory(c *gin.Context) {
	var history models.History
	if err := c.ShouldBindJSON(&history); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	query := `INSERT INTO history (id, user_id, event_id,
	 start_date, end_date, xp_earned, created_at, updated_at) 
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	_, err := h.db.Exec(query, history.ID, history.UserID,
		history.EventID, history.StartDate, history.EndDate,
		history.XPEarned, history.CreatedAt, history.UpdatedAt)
	if err != nil {
		log.Printf("Error creating history record: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating history record"})
		return
	}

	c.JSON(http.StatusCreated, history)
}

// @Summary     Get History
// @Description This API returns history details based on the provided history ID
// @Tags         History
// @Accept       json
// @Produce      json
// @Param        id   path string  true  "History ID"
// @Success      200  {object} models.History
// @Failure      400  {object} ErrorResponse
// @Failure      404  {object} ErrorResponse
// @Failure      500  {object} ErrorResponse
// @Router       /history/{id} [get]
func (h *HandlerV1) GetHistory(c *gin.Context) {
	id := c.Param("id")
	query := `SELECT id, user_id, event_id, start_date, end_date, xp_earned, created_at, updated_at 
				FROM history WHERE id = $1`

	var history models.History
	err := h.db.Get(&history, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "History record not found"})
		} else {
			log.Printf("Error fetching history record: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching history record"})
		}
		return
	}

	c.JSON(http.StatusOK, history)
}

// @Summary     Update History
// @Description This API updates an existing history record
// @Tags         History
// @Accept       json
// @Produce      json
// @Param        id   path string  true  "History ID"
// @Param        body body models.History true "Updated History Data"
// @Success      200  {object} models.History
// @Failure      400  {object} ErrorResponse
// @Failure      404  {object} ErrorResponse
// @Failure      500  {object} ErrorResponse
// @Router       /history/{id} [put]
func (h *HandlerV1) UpdateHistory(c *gin.Context) {
	id := c.Param("id")
	var history models.History
	if err := c.ShouldBindJSON(&history); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	query := `UPDATE history 
				SET user_id = $1, event_id = $2, start_date = $3, 
				end_date = $4, xp_earned = $5, updated_at = $6 
				WHERE id = $7`
	_, err := h.db.Exec(query, history.UserID, history.EventID, history.StartDate,
		history.EndDate, history.XPEarned, history.UpdatedAt, id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "History record not found"})
		} else {
			log.Printf("Error updating history record: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating history record"})
		}
		return
	}

	c.JSON(http.StatusOK, history)
}

// @Summary     Delete History
// @Description This API deletes a history record based on the provided history ID
// @Tags         History
// @Accept       json
// @Produce      json
// @Param        id   path string  true  "History ID"
// @Success      204  {object} nil
// @Failure      400  {object} ErrorResponse
// @Failure      404  {object} ErrorResponse
// @Failure      500  {object} ErrorResponse
// @Router       /history/{id} [delete]
func (h *HandlerV1) DeleteHistory(c *gin.Context) {
	id := c.Param("id")
	query := `DELETE FROM history WHERE id = $1`

	result, err := h.db.Exec(query, id)
	if err != nil {
		log.Printf("Error deleting history record: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting history record"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error checking rows affected: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking rows affected"})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "History record not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary     List History
// @Description This API returns a list of all history records
// @Tags         History
// @Accept       json
// @Produce      json
// @Success      200  {array} models.History
// @Failure      500  {object} ErrorResponse
// @Router       /history [get]
func (h *HandlerV1) ListHistory(c *gin.Context) {
	query := `SELECT id, user_id, event_id, start_date, end_date, 
				xp_earned, created_at, updated_at FROM history`

	var history []models.History
	err := h.db.Select(&history, query)
	if err != nil {
		log.Printf("Error fetching history records: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching history records"})
		return
	}

	c.JSON(http.StatusOK, history)
}

// CreateMarket creates a new market record
// @Summary     Create Market
// @Description This API creates a new market record
// @Tags         Market
// @Accept       json
// @Produce      json
// @Param        market body models.Market true "Market"
// @Success      201  {object} models.Market
// @Failure      500  {object} ErrorResponse
// @Router       /market [post]
func (h *HandlerV1) CreateMarket(c *gin.Context) {
	var market models.Market
	if err := c.ShouldBindJSON(&market); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	market.CreatedAt = time.Now()
	market.UpdatedAt = time.Now()

	query := `INSERT INTO market (name, description, count, xp, category_name, created_at, updated_at) 
              VALUES (:name, :description, :count, :xp, :category_name, :created_at, :updated_at) RETURNING id`
	stmt, err := h.db.PrepareNamed(query)
	if err != nil {
		log.Printf("Error preparing query: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error preparing query"})
		return
	}

	err = stmt.Get(&market.ID, market)
	if err != nil {
		log.Printf("Error inserting market record: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting market record"})
		return
	}

	c.JSON(http.StatusCreated, market)
}

// GetMarket retrieves a market record by ID
// @Summary     Get Market
// @Description This API retrieves a market record by ID
// @Tags         Market
// @Accept       json
// @Produce      json
// @Param        id path int true "Market ID"
// @Success      200  {object} models.Market
// @Failure      404  {object} ErrorResponse
// @Router       /market/{id} [get]
func (h *HandlerV1) GetMarket(c *gin.Context) {
	id := c.Param("id")
	var market models.Market

	query := `SELECT id, name, description, count, xp, category_name, created_at, updated_at FROM market WHERE id = $1`
	err := h.db.Get(&market, query, id)
	if err != nil {
		log.Printf("Error fetching market record: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Market record not found"})
		return
	}

	c.JSON(http.StatusOK, market)
}

// UpdateMarket updates an existing market record
// @Summary     Update Market
// @Description This API updates an existing market record
// @Tags         Market
// @Accept       json
// @Produce      json
// @Param        id path int true "Market ID"
// @Param        market body models.Market true "Market"
// @Success      200  {object} models.Market
// @Failure      500  {object} ErrorResponse
// @Router       /market/{id} [put]
func (h *HandlerV1) UpdateMarket(c *gin.Context) {
	id := c.Param("id")
	var market models.Market
	if err := c.ShouldBindJSON(&market); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	market.UpdatedAt = time.Now()

	query := `UPDATE market SET name = :name, description = :description, count = :count, xp = :xp, 
              category_name = :category_name, updated_at = :updated_at WHERE id = :id`
	stmt, err := h.db.PrepareNamed(query)
	if err != nil {
		log.Printf("Error preparing query: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error preparing query"})
		return
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		log.Printf("Error preparing query: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error preparing query"})
		return
	}
	market.ID = int64(idInt)
	_, err = stmt.Exec(market)
	if err != nil {
		log.Printf("Error updating market record: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating market record"})
		return
	}

	c.JSON(http.StatusOK, market)
}

// DeleteMarket deletes a market record by ID
// @Summary     Delete Market
// @Description This API deletes a market record by ID
// @Tags         Market
// @Accept       json
// @Produce      json
// @Param        id path int true "Market ID"
// @Success      204  {object} nil
// @Failure      500  {object} ErrorResponse
// @Router       /market/{id} [delete]
func (h *HandlerV1) DeleteMarket(c *gin.Context) {
	id := c.Param("id")

	query := `DELETE FROM market WHERE id = $1`
	_, err := h.db.Exec(query, id)
	if err != nil {
		log.Printf("Error deleting market record: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting market record"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListMarkets lists all market records
// @Summary     List Markets
// @Description This API lists all market records
// @Tags         Market
// @Accept       json
// @Produce      json
// @Success      200  {array} models.Market
// @Failure      500  {object} ErrorResponse
// @Router       /market [get]
func (h *HandlerV1) ListMarkets(c *gin.Context) {
	var markets []models.Market

	query := `SELECT id, name, description, count, xp, category_name, created_at, updated_at, image_url FROM market`
	err := h.db.Select(&markets, query)
	if err != nil {
		log.Printf("Error fetching market records: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching market records"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"products": markets})
}

// CheckUserXP checks if the user's XP is enough to buy an item from the market
// @Summary     Check User XP
// @Description This API checks if the user's XP is enough to buy an item from the market
// @Tags         Market
// @Accept       json
// @Produce      json
// @Param        userId path int true "User ID"
// @Param        itemId path int true "Item ID"
// @Success      200  {object} models.Message
// @Failure      400  {object} ErrorResponse
// @Failure      404  {object} ErrorResponse
// @Failure      500  {object} ErrorResponse
// @Router       /market/check/{userId}/{itemId} [get]
func (h *HandlerV1) CheckUserXP(c *gin.Context) {
	userId := c.Param("userId")
	itemId := c.Param("itemId")

	var user models.User
	userQuery := "SELECT xp FROM users WHERE id = $1"
	err := h.db.Get(&user, userQuery, userId)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user data"})
		}
		return
	}
	var item models.Market
	itemQuery := "SELECT xp FROM market WHERE id = $1"
	err = h.db.Get(&item, itemQuery, itemId)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching item data"})
		}
		return
	}

	if user.XP >= int(item.XP) {
		c.JSON(http.StatusOK, gin.H{"can_buy": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"can_buy": false})
	}
}

// OrderItem handles the order process for a user
// @Summary     Order Item
// @Description This API allows a user to order an item from the market if they have enough XP
// @Tags         Market
// @Accept       json
// @Produce      json
// @Param        userId path int true "User ID"
// @Param        itemId path int true "Item ID"
// @Success      200  {object} models.Message
// @Failure      400  {object} ErrorResponse
// @Failure      404  {object} ErrorResponse
// @Failure      500  {object} ErrorResponse
// @Router       /market/order/{userId}/{itemId} [post]
func (h *HandlerV1) OrderItem(c *gin.Context) {
	userId := c.Param("userId")
	itemId := c.Param("itemId")

	var user models.User
	userQuery := "SELECT id, xp FROM users WHERE id = $1"
	err := h.db.Get(&user, userQuery, userId)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user data"})
		}
		return
	}
	var item models.Market
	itemQuery := "SELECT id, xp FROM market WHERE id = $1"
	err = h.db.Get(&item, itemQuery, itemId)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching item data"})
		}
		return
	}

	if user.XP < int(item.XP) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not enough XP"})
		return
	}

	newXP := user.XP - int(item.XP)
	updateUserQuery := "UPDATE users SET xp = $1 WHERE id = $2"
	_, err = h.db.Exec(updateUserQuery, newXP, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating user XP"})
		return
	}
	orderNumber := rand.Intn(90000) + 10000 // Generates a number between 10000 and 99999

	order := models.Order{
		UserID:      user.ID,
		ItemID:      int(item.ID),
		OrderNumber: orderNumber,
		CreatedAt:   time.Now(),
	}
	orderQuery := `INSERT INTO orders (user_id, item_id, order_number, created_at) VALUES ($1, $2, $3, $4)`
	_, err = h.db.Exec(orderQuery, order.UserID, order.ItemID, order.OrderNumber, order.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order placed successfully", "order_number": orderNumber})
}
