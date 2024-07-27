package webhandlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"worker-bot/models"

	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"
	"github.com/jmoiron/sqlx"
	"github.com/k0kubun/pp"
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
// @Success     200 {object} ErrorResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /questions [get]
func (h *HandlerV1) TestGenHandler(c *gin.Context) {
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

	prompt := `{
  "tests": [
    {
      "1": "Which of the following is NOT a category within the broad field of Ecology?",
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
GENERATE ME 10 RANDOM ECOLOGY TESTS APPLYING THIS FORMAT ABOVE. QUESTION NUMBERS ARE DYNAMIC. QUESTIONS MUST BE VERY EASY AND FOR COMMON NATION. NOT FOR PROFESSORS.`

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
		pp.Println(err)
	}

	var carbine map[string]interface{}
	err = json.Unmarshal([]byte(unescapedData), &carbine)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error while unmarshaling answer to map",
		})

		return
	}

	c.JSON(http.StatusOK, carbine)

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
	err := h.db.Select(&users, "SELECT id, first_name, last_name, avatar, xp FROM users ORDER BY xp DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching data"})
		return
	}

	rankings := make([]models.RankingResponse, len(users))
	for i, user := range users {
		rankings[i] = models.RankingResponse{
			ID:       user.ID,
			Rank:     i + 1,
			UserName: user.FirstName + " " + user.LastName,
			XP:       user.XP,
			Avatar:   user.Avatar,
		}
	}

	c.JSON(http.StatusOK, rankings)
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
	query := `SELECT id, first_name, last_name, avatar, birth_date, location, phone_number, xp FROM users`

	var users []models.User
	err := h.db.Select(&users, query)
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching users"})
		return
	}

	c.JSON(http.StatusOK, users)
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
	query := `SELECT id, name, description, 
				total_xp, start_date, end_date, 
				resp_officer, resp_officer_image, 
				created_at, updated_at FROM events`

	var events []models.Event
	err := h.db.Select(&events, query)
	if err != nil {
		log.Printf("Error fetching events: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching events"})
		return
	}

	c.JSON(http.StatusOK, events)
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
