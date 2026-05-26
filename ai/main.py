from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import List
from openai import OpenAI
import os
import json
import re
from dotenv import load_dotenv


load_dotenv()
app = FastAPI()
print(os.getenv("GROQ_API_KEY"))

client = OpenAI(
    api_key=os.getenv("GROQ_API_KEY"),
    #base_url="https://api.groq.com/openai/v1",
    base_url="https://api.openai.com/v1",
)

MODEL_NAME = "llama-3.1-8b-instant"


class Grade(BaseModel):
    subject: str
    score: int


class StudentRequest(BaseModel):
    student_id: str
    student_name: str
    student_surname: str
    grades: List[Grade]


def extract_json(text: str):
    text = text.strip()

    text = re.sub(r"^```json\s*", "", text)
    text = re.sub(r"^```\s*", "", text)
    text = re.sub(r"\s*```$", "", text)

    start = text.find("{")
    end = text.rfind("}")

    if start != -1 and end != -1:
        text = text[start:end + 1]

    return json.loads(text)


@app.get("/")
def root():
    return {"status": "ok", "message": "FastAPI server is running"}


@app.post("/get_recommendations")
def get_recommendations(data: StudentRequest):
    if not os.getenv("GROQ_API_KEY"):
        raise HTTPException(
            status_code=500,
            detail="GROQ_API_KEY не найден."
        )

    grades_text = "\n".join(
        f"- {grade.subject}: {grade.score}" for grade in data.grades
    )

    prompt = prompt = f"""
Ты — опытный образовательный ИИ-помощник и персональный наставник.

Проанализируй оценки ученика и дай глубокие, персонализированные рекомендации на русском языке.

Данные ученика:
student_id: {data.student_id}
student_name: {data.student_name}
student_surname: {data.student_surname}

Оценки (шкала 0–100):
{grades_text}

Правила оценки уровня:
- 90–100: отличный уровень
- 75–89: хороший уровень
- 60–74: средний уровень, есть пробелы
- ниже 60: слабый уровень, нужна серьёзная работа

Для каждого предмета дай конкретную рекомендацию с учётом балла:
- Что именно делать (метод, техника, частота)
- Одну ссылку на бесплатный ресурс (Khan Academy, Coursera, YouTube-каналы, Stepik и т.д.)

Верни только JSON без markdown.

Формат:
{{
  "student_id": "{data.student_id}",
  "student_name": "{data.student_name}",
  "student_surname": "{data.student_surname}",
  "strengths": [
    "Конкретная сильная сторона с пояснением"
  ],
  "weaknesses": [
    "Конкретная слабая сторона с пояснением"
  ],
  "recommendations": [
    {{
      "subject": "Название предмета",
      "score": 0,
      "level": "отличный | хороший | средний | слабый",
      "recommendation": "Конкретный совет: что делать, как часто, какой метод использовать (2–3 предложения)",
      "action_steps": [
        "Шаг 1: конкретное действие на этой неделе",
        "Шаг 2: конкретное действие на следующей неделе",
        "Шаг 3: долгосрочная цель на месяц"
      ],
      "resources": [
        {{
          "title": "Название ресурса",
          "url": "https://...",
          "description": "Что именно там изучить"
        }}
      ]
    }}
  ],
  "general_advice": "Персонализированный совет, учитывающий общую картину успеваемости: баланс между предметами, режим учёбы, конкретные техники (метод Помодоро, интервальное повторение и т.д.)",
  "weekly_plan": "Краткий план на неделю: сколько времени уделить каждому предмету с учётом приоритетов"
}}
""".strip()

    try:
        completion = client.chat.completions.create(
            model=MODEL_NAME,
            messages=[
                {
                    "role": "system",
                    "content": "Ты возвращаешь только валидный JSON без markdown."
                },
                {
                    "role": "user",
                    "content": prompt
                }
            ],
            temperature=0.2,
                max_tokens=1500,
        )

        model_text = completion.choices[0].message.content

        try:
            return extract_json(model_text)
        except Exception:
            return {
                "student_id": data.student_id,
                "student_name": data.student_name,
                "student_surname": data.student_surname,
                "raw_recommendations": model_text
            }

    except Exception as e:
        print("GROQ ERROR:", repr(e))
        raise HTTPException(
            status_code=500,
            detail=f"Ошибка при обращении к Groq: {repr(e)}"
        )