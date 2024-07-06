package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"

	"github.com/joho/godotenv"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

type GetVacancies struct {
	Name    string
	Company string
	Salary  string
}

type VacancyAnalitics struct {
	Count              int    `json:"count"`
	MostCompany        string `json:"most_company"`
	CountOfMostCompany int    `json:"count_of_most_company"`
	AvaragePayment     int    `json:"avarage_payment"`
}

func telegramBot() {

	//Создаем бота
	if err := godotenv.Load(); err != nil {
		fmt.Println("Token error: no .env file")
	}
	token, ok := os.LookupEnv("TOKEN")
	if !ok {
		fmt.Println("Token error: no token in .env file")
	}
	fmt.Println(token)
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
	if err != nil {
		fmt.Println("Error while init bot")
	}

	//Устанавливаем время обновления
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	//Получаем обновления от бота
	updates, _ := bot.GetUpdatesChan(u)

	var qs []string

	for update := range updates {
		if update.Message == nil {
			continue
		}

		//Проверяем что от пользователья пришло именно текстовое сообщение
		if reflect.TypeOf(update.Message.Text).Kind() == reflect.String && update.Message.Text != "" {

			switch update.Message.Text {
			case "/start":

				//Отправлем сообщение
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, `Отправь сообщение формата: {должность};{компания};{минимальная зарплата}`)
				bot.Send(msg)

			case "/search_vacancies":
				//Sending GET query to 127.0.0.1:8000/vacancies
				url, _ := url.Parse("http://127.0.0.1:8000/search/vacancies")
				q := url.Query()
				fmt.Printf("%#v\n", qs)
				if qs[0] != "" {
					q.Add("company", qs[0])
				}
				if qs[1] != "" {
					q.Add("name", qs[1])
				}
				if qs[2] != "" {
					q.Add("salary", qs[2])
				}
				url.RawQuery = q.Encode()
				fmt.Println(url.String())

				resp, err := http.Get(url.String())
				if err != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Возникла ошибка")
					bot.Send(msg)
					continue
				}
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Возникла ошибка")
					bot.Send(msg)
					continue
				}
				var result VacancyAnalitics
				err = json.Unmarshal(body, &result)
				if err != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Возникла ошибка при десериализации")
					bot.Send(msg)
					continue
				}

				text := fmt.Sprintf("Найдено: %d вакансий\nБольше всего вакансий от - %s(%d вакансий)\nСредняя зарплата - %d",
					result.Count, result.MostCompany, result.CountOfMostCompany, result.AvaragePayment)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
				bot.Send(msg)
				//fmt.Println(result)

			default:
				qs = make([]string, 3)
				var s []rune
				var cnt int
				for _, x := range update.Message.Text {
					if x == ';' {
						if len(s) != 0 {
							qs[cnt] = string(s)
							s = nil
						}
						cnt += 1
					} else {
						s = append(s, x)
					}
				}
				qs[cnt] = string(s)
			}
		} else {

			//Отправлем сообщение
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Используйте слова.")
			bot.Send(msg)
		}
	}
}

func main() {
	telegramBot()
}
