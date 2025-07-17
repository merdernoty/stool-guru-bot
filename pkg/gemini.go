package gemini

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/genai"
)

type GeminiService struct {
	client *genai.Client
	model  *genai.Model
}

type AnalysisResult struct {
	Text        string `json:"text"`
	Diagnosis   string `json:"diagnosis"`
	Recommendations string `json:"recommendations"`
}

func NewGeminiService(apiKey string) (*GeminiService, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API ключ не может быть пустым")
	}

	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка создания клиента Gemini: %w", err)
	}

	model := client.GetModel("gemini-2.0-flash-exp")
	if model == nil {
		return nil, fmt.Errorf("не удалось получить модель Gemini")
	}

	log.Println("✅ Gemini сервис успешно инициализирован")

	return &GeminiService{
		client: client,
		model:  model,
	}, nil
}

func (g *GeminiService) AnalyzeImage(ctx context.Context, imageBytes []byte, mimeType string) (*AnalysisResult, error) {
	if len(imageBytes) == 0 {
		return nil, fmt.Errorf("данные изображения не могут быть пустыми")
	}

	if mimeType == "" {
		mimeType = "image/jpeg"
	}

	parts := []*genai.Part{
		{
			Text: `Проанализируй данное изображение стула/кала как опытный врач-гастроэнтеролог. 
			
			Пожалуйста, дай подробный анализ по следующим критериям:
			1. Форма и консистенция (согласно Бристольской шкале стула)
			2. Цвет и возможные причины изменений
			3. Размер и общий вид
			4. Потенциальные проблемы или отклонения от нормы
			5. Конкретные рекомендации по питанию
			6. Когда следует обратиться к врачу
			
			Ответ структурируй в формате:
			ДИАГНОЗ: [краткая оценка состояния]
			РЕКОМЕНДАЦИИ: [конкретные советы по питанию и образу жизни]
			ВНИМАНИЕ: [когда нужна медицинская помощь]`,
		},
		{
			InlineData: &genai.Blob{
				Data:     imageBytes,
				MIMEType: mimeType,
			},
		},
	}

	result, err := g.client.Models.GenerateContent(ctx, g.model.Name,
		genai.NewUserContent(parts...),
		&genai.GenerateContentConfig{
			Temperature:     0.7,
			MaxOutputTokens: 1000,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("ошибка генерации контента: %w", err)
	}

	if result == nil || result.Text() == "" {
		return nil, fmt.Errorf("получен пустой ответ от Gemini")
	}

	log.Printf("🔬 Анализ изображения завершен, длина ответа: %d символов", len(result.Text()))

	analysisResult := &AnalysisResult{
		Text: result.Text(),
	}

	analysisResult.Diagnosis, analysisResult.Recommendations = g.parseResponse(result.Text())

	return analysisResult, nil
}

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
		{Text: prompt},
		{
			InlineData: &genai.Blob{
				Data:     imageBytes,
				MIMEType: mimeType,
			},
		},
	}

	result, err := g.client.Models.GenerateContent(ctx, g.model.Name,
		genai.NewUserContent(parts...),
		&genai.GenerateContentConfig{
			Temperature:     0.7,
			MaxOutputTokens: 1000,
		},
	)

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

func (g *GeminiService) parseResponse(response string) (diagnosis, recommendations string) {
	lines := []rune(response)
	text := string(lines)
		
	if len(text) > 100 {
		diagnosis = text[:100] + "..."
		if len(text) > 200 {
			recommendations = text[100:200] + "..."
		}
	} else {
		diagnosis = text
	}
	
	return diagnosis, recommendations
}

func (g *GeminiService) SendTextMessage(ctx context.Context, message string) (*AnalysisResult, error) {
	if message == "" {
		return nil, fmt.Errorf("сообщение не может быть пустым")
	}

	parts := []*genai.Part{
		{Text: message},
	}

	result, err := g.client.Models.GenerateContent(ctx, g.model.Name,
		genai.NewUserContent(parts...),
		&genai.GenerateContentConfig{
			Temperature:     0.7,
			MaxOutputTokens: 500,
		},
	)

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

func (g *GeminiService) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}
func (g *GeminiService) GetModelInfo() string {
	if g.model != nil {
		return g.model.Name
	}
	return "модель не загружена"
}