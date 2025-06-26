package utils

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"
)

func AfterNow(now time.Time, date time.Time) bool {
	return date.After(now)
}

func SplitRule(repeat string) []string {
	repeat = strings.ReplaceAll(repeat, ",", " ")
	return strings.Split(repeat, " ")
}

func DayRule(now time.Time, start time.Time, rule string) (string, error) {
	var date time.Time
	repeat := SplitRule(rule)
	if len(repeat) != 2 {
		return "", errors.New("некорректный ввод правила, введите в формате: 'd number'")
	}
	day, err := strconv.Atoi(repeat[1])
	if err != nil {
		return "", errors.New("ошибка конвертации string --> int")
	}
	if day < 1 || day > 400 {
		return "", errors.New("некорректный ввод количества дней. Укажите дни в диапазоне [1:400]")
	}
	for {
		date = start.AddDate(0, 0, day)
		if AfterNow(now, date) {
			break
		}
		start = date
	}
	return date.Format(DateFormat), nil
}

func WeekRule(now time.Time, start time.Time, rule string) (string, error) {
	var date time.Time
	repeat := SplitRule(rule)
	if len(repeat) < 2 {
		return "", errors.New("некорректный ввод дней недели. Укажите дни в формате 'w number (number, ...)'")
	}
	minDays := 8

	currentDay := weekdayToInt(now.Weekday())
	for _, dayStr := range repeat[1:] {
		day, err := strconv.Atoi(dayStr)
		if err != nil {
			return "", errors.New("ошибка конвертации string --> int")
		}
		if day < 1 || day > 7 {
			return "", errors.New("дни недели в неверном диапазоне [1:7]")
		}
		var daysLeft int
		if day == currentDay {
			daysLeft = 7 // Следующая неделя
		} else if day < currentDay {
			daysLeft = 7 - (currentDay - day)
		} else {
			daysLeft = day - currentDay
		}

		if daysLeft < minDays {
			minDays = daysLeft
		}
	}

	date = now.AddDate(0, 0, minDays)

	return date.Format(DateFormat), nil
}

func weekdayToInt(w time.Weekday) int {
	if w == 0 {
		return 7 // воскресенье — 7
	}
	return int(w)
}

func MonthRule(now time.Time, start time.Time, rule string) (string, error) {
	ruleFields := strings.Fields(rule)
	if len(ruleFields) < 2 {
		return "", errors.New("не верный формат ввода 'm num(num,...) startMonth (startMonth,...)'")
	}

	var days []int
	var months []int

	for _, s := range strings.Split(ruleFields[1], ",") {
		day, err := strconv.Atoi(s)
		if err != nil {
			return "", errors.New("ошибка конвертации string -> int (день)")
		}
		if day < -2 || day == 0 || day > 31 {
			return "", errors.New("не верный день [-2,-1, 1...31]")
		}
		days = append(days, day)
	}

	// Парсим месяцы (если указаны)
	if len(ruleFields) == 3 {
		for _, s := range strings.Split(ruleFields[2], ",") {
			month, err := strconv.Atoi(s)
			if err != nil {
				return "", errors.New("ошибка конвертации string -> int (месяц)")
			}
			if month < 1 || month > 12 {
				return "", errors.New("месяц должен быть в диапазоне [1:12]")
			}
			months = append(months, month)
		}
	} else {
		for i := 1; i <= 12; i++ {
			months = append(months, i)
		}
	}
	var earliest *time.Time
	month := months[0]

	slices.Sort(days)
	slices.Sort(months)

	if len(ruleFields) == 2 {

		month = int(now.Month())
		year := now.Year()
		for i := month; i <= 12; i++ {
			for _, day := range days {
				validDay, err := resolveDay(year, time.Month(i), day)
				if err != nil {
					continue // день невалиден для месяца
				}
				candidate := time.Date(year, time.Month(i), validDay, 0, 0, 0, 0, time.UTC)
				if candidate.After(start) && candidate.After(now) {
					if earliest == nil || candidate.Before(*earliest) {
						earliest = &candidate
					}
				}
				if earliest != nil && candidate.Year() > now.Year() {
					break // нашли дату в будущем, выходим из цикла
				}
				if i == 12 && earliest == nil {
					// если не нашли ни одной даты в будущем, начинаем с января следующего года
					year++
					i = 0 // сбрасываем месяц на январь
				}

			}
		}
	} else if len(ruleFields) == 3 {
		// Если указаны месяцы, начинаем с первого указанного месяца
		for y := now.Year(); y <= now.Year()+3; y++ {
			for _, month = range months {
				for _, day := range days {
					validDay, err := resolveDay(y, time.Month(month), day)
					if err != nil {
						continue // день невалиден для месяца
					}
					candidate := time.Date(y, time.Month(month), validDay, 0, 0, 0, 0, time.UTC)
					if candidate.After(start) && candidate.After(now) {
						if earliest == nil || candidate.Before(*earliest) {
							earliest = &candidate
						}
					}
				}
				if earliest != nil {
					break // нашли дату в будущем, выходим из цикла
				}
			}
			if earliest != nil {
				break // нашли дату в будущем, выходим из цикла
			}
		}
	}
	if earliest == nil {
		return "", errors.New("нет доступных дат в будущем по указанным правилам")
	}
	return earliest.Format("20060102"), nil
}

func resolveDay(year int, month time.Month, day int) (int, error) {
	last := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	switch day {
	case -1:
		return last, nil
	case -2:
		if last > 1 {
			return last - 1, nil
		}
		return 0, errors.New("нет предпоследнего дня")
	default:
		if day >= 1 && day <= last {
			return day, nil
		}
		return 0, errors.New("указанный день не существует в этом месяце")
	}
}

func YearRule(now time.Time, start time.Time, rule string) (string, error) {
	var date time.Time
	repeat := SplitRule(rule)
	if len(repeat) != 1 {
		return "", errors.New("некорректный ввод правила, введите в формате: 'y'")
	}
	for {
		date = start.AddDate(1, 0, 0)
		if AfterNow(now, date) {
			break
		}
		start = date
	}
	return date.Format(DateFormat), nil
}

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if len(repeat) == 0 {
		return "", errors.New("правило повторения не указано")
	}
	runes := []rune(repeat)
	valid := map[rune]bool{
		'd': true,
		'w': true,
		'm': true,
		'y': true,
	}
	if !valid[runes[0]] {
		return "", errors.New(`Некорректный ввод правила повторения. Пожалуйста укажите:
		'd' - для правила "день"
		'w' - для правила "неделя"
		'm' - для правила "месяц"
		'y' в случае правила "год"`)
	}

	start, err := time.Parse(DateFormat, dstart)
	if err != nil {
		return "", errors.New("некорректный ввод даты. Введите в формате 'YYYYMMDD'")
	}

	switch runes[0] {
	case 'd':
		return DayRule(now, start, repeat)
	case 'w':
		return WeekRule(now, start, repeat)
	case 'm':
		return MonthRule(now, start, repeat)
	case 'y':
		return YearRule(now, start, repeat)
	default:
		return "", errors.New("неизвестная ошибка при обработке правила")
	}

}
