package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
	PATTERN

	{
    "data": "",
    "datajuliano": "",
    "comemorados": "",
    "Fonte": "",
    "santos": [
        {
			"id": number
            "nome": "",
            "imagemurl": "",
            "conteudo": ""
        }
    ]
}

*/

type Santos struct {
	ID        int    `json:"id"`
	Nome      string `json:"nome"`
	ImagemUrl string `json:"imagemurl"`
	Conteudo  string `json:"conteudo"`
}

type Sinaxario struct {
	ID          int64    `json:"id"`
	Data        string   `json:"data"`
	DataJuliano string   `json:"datajuliano"`
	Comemorados string   `json:"comemorados"`
	Fonte       string   `json:"fonte"`
	Santos      []Santos `json:"santos"`
}

var client *mongo.Client

func setupCollection() *mongo.Collection {
	return client.Database("Sinaxario").Collection("Sinaxario")
}

func init() {
	err := godotenv.Load(".env")

	if err != nil {
		panic(err)
	}

	dbURL := os.Getenv("DB_URL")

	dbAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(dbURL).SetServerAPIOptions(dbAPI)

	client, err = mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	if err := client.Ping(context.TODO(), nil); err != nil {
		panic(err)
	}

	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
}

func add(c *gin.Context) {
	var newData Sinaxario

	if err := c.BindJSON(&newData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Um Erro foi encontrado: ": err.Error()})
		return
	}

	collection := setupCollection()
	collectionSize, err := collection.CountDocuments(context.TODO(), bson.D{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newData.ID = collectionSize + 1

	result, err := collection.InsertOne(context.TODO(), newData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"result": result.InsertedID})
}

func getAll(c *gin.Context) {
	filter := bson.D{}
	collection := setupCollection()
	cursor, err := collection.Find(context.TODO(), filter)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(context.TODO())

	var results []Sinaxario
	if err := cursor.All(context.TODO(), &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"result": results})
}

func finByData(c *gin.Context) {
	data := c.Query("data")
	collection := setupCollection()
	filter := bson.D{{"data", data}}

	var result Sinaxario

	err := collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"erro": "Dados com a data providenciada inexistente."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, result)
}

func main() {
	router := gin.Default()
	router.POST("/sinaxario", add)
	router.GET("/sinaxario", getAll)
	router.GET("/sinaxario/findby", finByData)

	router.Run("localhost:8080")
}
