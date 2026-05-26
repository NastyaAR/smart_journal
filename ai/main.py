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

client = OpenAI(
    api_key=os.getenv("OPENROUTER_API_KEY"),
    base_url="https://openrouter.ai/api/v1",
)

MODEL_NAME = "meta-llama/llama-3.1-8b-instruct"


# -------------------------
# MODELS
# -------------------------

class Grade(BaseModel):
    subject: str
    score: int


class StudentRequest(BaseModel):
    student_id: str
    student_name: str
    student_surname: str
    grades: List[Grade]


# -------------------------
# JSON PARSER
# -------------------------

def extract_json(text: str):
    text = text.strip()
    text = re.sub(r"^json\s*", "", text, flags=re.IGNORECASE)
    text = text.strip("```")

    start = text.find("{")
    end = text.rfind("}")

    if start != -1 and end != -1:
        text = text[start:end + 1]

    return json.loads(text)


# -------------------------
# NORMALIZER (FIX EMPTY FIELDS)
# -------------------------

def normalize_output(data: dict, grades: List[Grade]):
    if not data.get("strengths"):
        data["strengths"] = ["Стабильные результаты в ряде предметов"]

    if not data.get("weaknesses"):
        data["weaknesses"] = ["Некоторые предметы требуют дополнительной практики"]

    if not data.get("recommendations"):
        data["recommendations"] = [
            {
                "subject": g.subject,
                "score": g.score,
                "recommendation": "Рекомендуется больше практики и повторения материала"
            }
            for g in grades[:3]
        ]

    if not data.get("general_advice"):
        data["general_advice"] = (
            "Регулярно повторяйте материал и уделяйте внимание слабым предметам"
        )

    return data


# -------------------------
# ROUTES
# -------------------------

@app.get("/")
def root():
    return {"status": "ok", "message": "FastAPI server is running"}


@app.post("/get_recommendations")
def get_recommendations(data: StudentRequest):

    if not os.getenv("OPENROUTER_API_KEY"):
        raise HTTPException(
            status_code=500,
            detail="OPENROUTER_API_KEY не найден"
        )

    grades_text = "\n".join(
        f"- {g.subject}: {g.score}" for g in data.grades
    )

    prompt = f"""
Ты — опытный образовательный аналитик.

Проанализируй оценки ученика и выдай структурированный ответ.

ДАННЫЕ:
ID: {data.student_id}
Имя: {data.student_name} {data.student_surname}

Оценки:
{grades_text}

ПРАВИЛА АНАЛИЗА:
- 90–100: отличный уровень
- 75–89: хороший уровень
- 60–74: средний уровень
- ниже 60: слабый уровень

ЖЕСТКИЕ ТРЕБОВАНИЯ:
- strengths: минимум 2 пункта
- weaknesses: минимум 2 пункта
- recommendations: минимум 3 пункта
- НЕЛЬЗЯ писать "нет данных" или пустые значения
- Если данных мало — делай логический вывод
- Если оценка больше 50, то нельзя писать, что она низкая

ФОРМАТ ОТВЕТА:
Верни ТОЛЬКО валидный JSON:

{{
  "student_id": "{data.student_id}",
  "student_name": "{data.student_name}",
  "student_surname": "{data.student_surname}",
  "strengths": [],
  "weaknesses": [],
  "recommendations": [
    {{
      "subject": "пример",
      "score": 0,
      "recommendation": "совет"
    }}
  ],
  "general_advice": ""
}}
""".strip()

    try:
        completion = client.chat.completions.create(
            model=MODEL_NAME,
            messages=[
                {"role": "system", "content": "Ты возвращаешь только валидный JSON."},
                {"role": "user", "content": prompt}
            ],
            temperature=0.7,
            max_tokens=1500,
        )

        model_text = completion.choices[0].message.content

        try:
            result = extract_json(model_text)
            return normalize_output(result, data.grades)

        except Exception:
            return normalize_output(
                {
                    "student_id": data.student_id,
                    "student_name": data.student_name,
                    "student_surname": data.student_surname,
                    "strengths": [],
                    "weaknesses": [],
                    "recommendations": [],
                    "general_advice": ""
                },
                data.grades
            )

    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"Ошибка OpenRouter: {repr(e)}"
        )