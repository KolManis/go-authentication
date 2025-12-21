package helper

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/KolManis/go-authentication/database"
	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Email      string
	First_name string
	Last_name  string
	Uid        string
	User_type  string
	jwt.StandardClaims
}

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var SECRET_KEY string = os.Getenv("SECRET_KEY")

func GenerateAllTokens(
	email string,
	firstName string,
	lastName string,
	userType string,
	uid string,
) (signedtoken string, singnedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email:      email,
		First_name: firstName,
		Last_name:  lastName,
		Uid:        uid,
		User_type:  userType,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodES256, claims).SignedString([]byte(SECRET_KEY))
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodES256, refreshClaims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return
	}

	return token, refreshToken, err
}

func VaildateToken(signedToken string) (claims *SignedDetails, msg string) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)
	if err != nil {
		msg = err.Error()
		return
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = fmt.Sprintf("the token is invalid")
		msg = err.Error()
		return
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = fmt.Sprintf("the token is expired")
		msg = err.Error()
		return
	}

	return claims, msg
}

func UpdateAllTokens(signedtoken string, signedRefreshToken string, userId string) error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var updateObj primitive.D

	updateObj = append(updateObj, bson.E{"token", signedtoken})
	updateObj = append(updateObj, bson.E{"refresh_token", signedRefreshToken})

	Update_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{"update_at", Update_at})

	// primitive.D{
	// {"token", "eyJhbGciOiJ..."},
	// {"refresh_token", "eyJhbGciOiJ..."},
	// {"update_at", "2024-01-15T10:30:00Z"},

	upsert := true
	filter := bson.M{"user_id": userId}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := userCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{"$set", updateObj},
		},
		&opt,
	)

	if err != nil {
		log.Panic(err)
		return err
	}

	return nil
}

// func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string) error {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	update := bson.D{
// 		{"$set", bson.D{
// 			{"token", signedToken},
// 			{"refresh_token", signedRefreshToken},
// 			{"updated_at", time.Now()},
// 		}},
// 	}

// 	filter := bson.M{"user_id": userId}

// 	// Опции (без upsert - пользователь должен существовать)
// 	opt := options.Update().SetUpsert(false)

// 	result, err := userCollection.UpdateOne(ctx, filter, update, opt)
// 	if err != nil {
// 		return fmt.Errorf("failed to update tokens: %v", err)
// 	}

// 	if result.MatchedCount == 0 {
// 		return errors.New("user not found")
// 	}

// 	return nil
// }
