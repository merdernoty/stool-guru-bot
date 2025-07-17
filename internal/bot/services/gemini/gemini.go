package gemini

import (
	"context"
	"fmt"
	"log"
	"strings"

	"google.golang.org/genai"
)

// GeminiService - сервис для работы с Gemini AI
type GeminiService struct {
	client *genai.Client
	model  string
}

// AnalysisResult - результат анализа изображения
type AnalysisResult struct {
	Text            string `json:"text"`
	Diagnosis       string `json:"diagnosis"`
	Recommendations string `json:"recommendations"`
}

// NewGeminiService создает новый экземпляр сервиса Gemini
func NewGeminiService(apiKey string) (*GeminiService, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API ключ не может быть пустым")
	}

	ctx := context.Background()

	// Используем новый API для создания клиента
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка создания клиента Gemini: %w", err)
	}

	log.Println("✅ Gemini сервис успешно инициализирован")

	return &GeminiService{
		client: client,
		model:  "gemini-2.0-flash", // Используем актуальную модель
	}, nil
}

// AnalyzeImage анализирует изображение с помощью Gemini AI
func (g *GeminiService) AnalyzeImage(ctx context.Context, imageBytes []byte, mimeType string) (*AnalysisResult, error) {
	if len(imageBytes) == 0 {
		return nil, fmt.Errorf("данные изображения не могут быть пустыми")
	}

	if mimeType == "" {
		mimeType = "image/jpeg" // значение по умолчанию
	}

	// Создаем части сообщения с текстом и изображением
	parts := []*genai.Part{
		genai.NewPartFromText(`Ты опытный врач-гастроэнтеролог. Проанализируй данное изображение стула/кала и дай профессиональную медицинскую оценку.

ВАЖНО: Анализируй только если на изображении действительно стул/кал. Если это что-то другое, скажи об этом.

Если это стул, дай анализ по следующим критериям:
1. ФОРМА И КОНСИСТЕНЦИЯ (по Бристольской шкале стула 1-7)
2. ЦВЕТ и возможные причины
3. РАЗМЕР и общий вид
4. ПОТЕНЦИАЛЬНЫЕ ПРОБЛЕМЫ

Структурируй ответ так:
🔬 АНАЛИЗ:
[Подробное описание]

⚕️ ОЦЕНКА:
[Оценка по Бристольской шкале и общее состояние]

🥗 РЕКОМЕНДАЦИИ:
[Конкретные советы по питанию]

⚠️ ВНИМАНИЕ:
[Когда нужна медицинская помощь]

Отвечай профессионально, но понятно. Напоминай, что это не заменяет консультацию врача.`),
		genai.NewPartFromBytes(imageBytes, mimeType),
	}

	// Создаем контент для передачи в модель
	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	// Генерируем контент с новым API
	result, err := g.client.Models.GenerateContent(ctx, g.model, contents, &genai.GenerateContentConfig{
		Temperature:     genai.Ptr(float32(0.7)),
		MaxOutputTokens: 1500,
		TopP:            genai.Ptr(float32(0.9)),
	})

	if err != nil {
		return nil, fmt.Errorf("ошибка генерации контента: %w", err)
	}

	if result == nil || result.Text() == "" {
		return nil, fmt.Errorf("получен пустой ответ от Gemini")
	}

	log.Printf("🔬 Анализ изображения завершен, длина ответа: %d символов", len(result.Text()))

	// Парсим ответ для структурированного результата
	analysisResult := &AnalysisResult{
		Text: result.Text(),
	}

	// Простой парсинг структурированного ответа
	analysisResult.Diagnosis, analysisResult.Recommendations = g.parseResponse(result.Text())

	return analysisResult, nil
}

// AnalyzeImageWithCustomPrompt анализирует изображение с кастомным промптом
func (g *GeminiService) AnalyzeImageWithCustomPrompt(ctx context.Context, imageBytes []byte, prompt string, mimeType string) (*AnalysisResult, error) {
	if len(imageBytes) == 0 {
		return nil, fmt.Errorf("данные изображения не могут быть пустыми")
	}

	if prompt == "" {
		return nil, fmt.Errorf("промпт не может быть пустым")
	}

	if mimeType == "" {
		mimeType = "image/jpeg"
	}

	parts := []*genai.Part{
		genai.NewPartFromText(prompt),
		genai.NewPartFromBytes(imageBytes, mimeType),
	}

	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	result, err := g.client.Models.GenerateContent(ctx, g.model, contents, &genai.GenerateContentConfig{
		Temperature:     genai.Ptr(float32(0.7)),
		MaxOutputTokens: 1500,
	})

	if err != nil {
		return nil, fmt.Errorf("ошибка генерации контента: %w", err)
	}

	if result == nil || result.Text() == "" {
		return nil, fmt.Errorf("получен пустой ответ от Gemini")
	}

	return &AnalysisResult{
		Text: result.Text(),
	}, nil
}

// parseResponse парсит ответ от Gemini для извлечения структурированной информации
func (g *GeminiService) parseResponse(response string) (diagnosis, recommendations string) {
	lines := strings.Split(response, "\n")
	
	var diagnosisLines []string
	var recommendationLines []string
	var currentSection string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.Contains(line, "АНАЛИЗ:") || strings.Contains(line, "ОЦЕНКА:") {
			currentSection = "diagnosis"
			continue
		} else if strings.Contains(line, "РЕКОМЕНДАЦИИ:") {
			currentSection = "recommendations"
			continue
		} else if strings.Contains(line, "ВНИМАНИЕ:") {
			currentSection = "attention"
			continue
		}
		
		if currentSection == "diagnosis" && line != "" {
			diagnosisLines = append(diagnosisLines, line)
		} else if currentSection == "recommendations" && line != "" {
			recommendationLines = append(recommendationLines, line)
		}
	}
	
	diagnosis = strings.Join(diagnosisLines, " ")
	recommendations = strings.Join(recommendationLines, " ")
	
	// Ограничиваем длину
	if len(diagnosis) > 200 {
		diagnosis = diagnosis[:200] + "..."
	}
	if len(recommendations) > 200 {
		recommendations = recommendations[:200] + "..."
	}
	
	return diagnosis, recommendations
}

// SendTextMessage отправляет текстовое сообщение в Gemini
func (g *GeminiService) SendTextMessage(ctx context.Context, message string) (*AnalysisResult, error) {
	if message == "" {
		return nil, fmt.Errorf("сообщение не может быть пустым")
	}

	// Используем удобную функцию genai.Text для создания контента
	contents := genai.Text(message)

	result, err := g.client.Models.GenerateContent(ctx, g.model, contents, &genai.GenerateContentConfig{
		Temperature:     genai.Ptr(float32(0.7)),
		MaxOutputTokens: 500,
	})

	if err != nil {
		return nil, fmt.Errorf("ошибка отправки текстового сообщения: %w", err)
	}

	if result == nil || result.Text() == "" {
		return nil, fmt.Errorf("получен пустой ответ от Gemini")
	}

	return &AnalysisResult{
		Text: result.Text(),
	}, nil
}

// Close закрывает соединение с клиентом Gemini
func (g *GeminiService) Close() error {
	// В новом API нет метода Close для клиента
	// Клиент автоматически управляет ресурсами
	return nil
}

// GetModelInfo возвращает информацию о модели
func (g *GeminiService) GetModelInfo() string {
	return g.model
}

// HealthCheck проверяет состояние сервиса
func (g *GeminiService) HealthCheck(ctx context.Context) error {
	if g.client == nil {
		return fmt.Errorf("клиент Gemini не инициализирован")
	}
	
	// Простая проверка с минимальным запросом
	_, err := g.SendTextMessage(ctx, "test")
	if err != nil {
		return fmt.Errorf("сервис Gemini недоступен: %w", err)
	}
	
	return nil
}