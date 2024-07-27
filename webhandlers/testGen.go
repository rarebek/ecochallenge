package webhandlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

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

type User struct {
	ID        int    `db:"id" json:"id"`
	FirstName string `db:"first_name" json:"first_name"`
	LastName  string `db:"last_name" json:"last_name"`
	Avatar    string `db:"avatar" json:"avatar"`
	XP        int    `db:"xp" json:"xp"`
}

type RankingResponse struct {
	Rank     int    `json:"rank"`
	UserName string `json:"user_name"`
	XP       int    `json:"xp"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type Response map[string]interface{}

// @Summary     Get Questions
// @Description This API generates 10 questions with Gemini AI providing their answers too.
// @Tags  	    question
// @Accept      json
// @Produce     json
// @Success     200 {object} ErrorResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /question/get [get]
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
// @Tags  	    ranking
// @Accept      json
// @Produce     json
// @Success     200 {object} []RankingResponse
// @Failure     400 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /ranking [get]
func (h *HandlerV1) GetRanking(c *gin.Context) {
	var users []User
	err := h.db.Select(&users, "SELECT id, first_name, last_name, avatar, xp FROM users ORDER BY xp DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching data"})
		return
	}

	rankings := make([]RankingResponse, len(users))
	for i, user := range users {
		rankings[i] = RankingResponse{
			Rank:     i + 1,
			UserName: user.FirstName + " " + user.LastName,
			XP:       user.XP,
		}
	}

	c.JSON(http.StatusOK, rankings)
}
