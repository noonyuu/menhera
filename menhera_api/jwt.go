package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type Foo struct {
	Token string `json:"access_token"`
	Type  string `json:"token_type"`
	Exp   int64  `json:"expires_in"`
	Id    string `json:"key_id"`
}

func Jwt() (string, error){
	// 環境変数ファイルの読み込み
	godotenv.Load(".env")

	// 秘密鍵のファイルを開く
	f, err := os.Open("private.key")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// ファイルから秘密鍵の読み込み
	b, err := ioutil.ReadAll(f)
	privkey, err := jwk.ParseKey(b)
	if err != nil {
		fmt.Printf("failed to parse JWK: %s\n", err)
		return "" ,err
	}

	{
		// audプロパティに追加するために、aud変数を作成
		var aud []string
		aud = append(aud, "https://api.line.me/") // audプロパティの値を追加

		// JWTを構成する
		tok, err := jwt.NewBuilder().
			Subject(os.Getenv("CHID")).
			Issuer(os.Getenv("CHID")).
			Audience(aud).
			Expiration(time.Now().Add(30 * time.Minute)).	// 有効期限を30分に設定
			Build()
		if err != nil {
			fmt.Printf("failed to build token: %s\n", err)
			return "" ,err
		}

		// token_expプロパティはメソッドが用意されてないので、.Setで追加。
		tok.Set("token_exp", 60*60*24*30) // token_expプロパティ、チャネルアクセストークンの有効期間を指定

		// JWTを発行する
		signed, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256, privkey)) // signedにJWTがエンコードされ代入される
		if err != nil {
			fmt.Printf("failed to sign token: %s\n", err)
			return "" ,err
		}

		form := url.Values{}
		form.Set("grant_type", "client_credentials")
		form.Add("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
		form.Add("client_assertion", string(signed))

		body := strings.NewReader(form.Encode()) // リクエストのbodyを作成

		// リクエストの作成
		req, err := http.NewRequest(http.MethodPost, "https://api.line.me/oauth2/v2.1/token", body)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// 作成したリクエストの送信
		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()

		// レスポンスの解析
		var r io.Reader = res.Body

		var foo Foo
		err = json.NewDecoder(r).Decode(&foo)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("チャネルアクセストークン")
		return foo.Token, nil
	}
}
