package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

type Bank struct {
	Name    string
	BinFrom int
	BinTo   int
}

const UnknownBank = "Неизвестный банк"

// extractDigits возвращает только цифры из строки
func extractDigits(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, s)
}

func loadBankData(path string) ([]Bank, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var banks []Bank
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		fields := strings.Split(line, ",")
		if len(fields) < 3 {
			return nil, fmt.Errorf("invalid line format: %q", fields)
		}

		binFrom, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, fmt.Errorf("parse binFrom: %w", err)
		}

		binTo, err := strconv.Atoi(fields[2])
		if err != nil {
			return nil, fmt.Errorf("parse binTo: %w", err)
		}

		if binFrom > binTo {
			return nil, fmt.Errorf("invalid bank range in line %q: binFrom (%d) > binTo (%d)", line, binFrom, binTo)
		}

		resultBank := Bank{
			Name:    fields[0],
			BinFrom: binFrom,
			BinTo:   binTo,
		}

		banks = append(banks, resultBank)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading: %w", err)
	}

	return banks, nil
}

func extractBIN(cardDigits string) (int, error) {
	if len(cardDigits) < 6 {
		return 0, fmt.Errorf("card number too short: got %d digits, need at least 6", len(cardDigits))
	}

	bin, _ := strconv.Atoi(cardDigits[:6])

	return bin, nil
}

func identifyBank(bin int, banks []Bank) string {
	for _, bank := range banks {
		if bin >= bank.BinFrom && bin <= bank.BinTo {
			return bank.Name
		}
	}

	return UnknownBank
}

func validateLuhn(cardDigits string) bool {
	if len(cardDigits) == 0 {
		return false
	}

	//Определяем, нужно ли удваивать первую цифру
	double := len(cardDigits)%2 == 0
	sum := 0
	for _, r := range cardDigits {
		digit := int(r - '0')
		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		double = !double
	}

	return sum%10 == 0
}

func getUserInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Введите номер карты (или нажмите Enter для выхода из программы): ")

	cardNumber, err := reader.ReadString('\n')
	if err == io.EOF {
		return "", err
	}

	return strings.TrimSpace(cardNumber), nil
}

func main() {
	fmt.Println("Добро пожаловать в программу валидации карт!")

	banks, err := loadBankData("banks.txt")
	if err != nil {
		log.Fatal(err)
	}

	for {
		rawInput, err := getUserInput()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("Ошибка чтения ввода: %v", err)
		}

		if rawInput == "" {
			fmt.Println("Программа завершена.")
			break
		}

		cardDigits := extractDigits(rawInput)
		digitCount := len(cardDigits)

		if digitCount < 13 || digitCount > 19 {
			fmt.Printf("Неверная длина номера: %d цифр (должно быть 13–19)\n", digitCount)
			continue
		}

		if !validateLuhn(cardDigits) {
			fmt.Println("Невалидный номер карты (не проходит проверку Луна)")
			continue
		}
		fmt.Println("✅ Номер карты валиден")

		bin, err := extractBIN(cardDigits)
		if err != nil {
			fmt.Printf("Ошибка при извлечении BIN: %v\n", err)
			continue
		}

		bankName := identifyBank(bin, banks)
		if bankName == UnknownBank {
			fmt.Println("Эмитент не определен")
			continue
		}

		fmt.Printf("Банк: %s\n", bankName)
	}
}
