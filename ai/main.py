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
MODEL_NAME = "mistralai/mistral-7b-instruct-v0.1"

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
    text = re.sub(r"^json\s*", "", text)
    text = re.sub(r"^\s*", "", text)
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
    api_key = os.getenv("OPENROUTER_API_KEY")
    if not api_key:
        raise HTTPException(
            status_code=500,
            detail="OPENROUTER_API_KEY не найден."
        )
    grades_text = "\n".join(
        f"- {grade.subject}: {grade.score}" for grade in data.grades
    )
    prompt = f"""
Ты — опытный образовательный ИИ-помощник.
Проанализируй оценки и дай рекомендации на русском языке.
Данные:
- student_id: {data.student_id}
- student_name: {data.student_name} {data.student_surname}
- Оценки: {grades_text}
Правила:
- 90-100: отличный уровень
- 75-89: хороший уровень  
- 60-74: средний уровень
- ниже 60: слабый уровень
Верни ТОЛЬКО JSON без markdown:
{{
  "student_id": "{data.student_id}",
  "student_name": "{data.student_name}",
  "student_surname": "{data.student_surname}",
  "strengths": ["сильная сторона"],
  "weaknesses": ["слабая сторона"],
  "recommendations": [
    {{"subject": "Предмет", "score": 85, "recommendation": "совет"}}
  ],
  "general_advice": "общий совет"
}}
""".strip()
    
    try:
        completion = client.chat.completions.create(
            model=MODEL_NAME,
            messages=[
                {"role": "system", "content": "Верни только JSON."},
                {"role": "user", "content": prompt}
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
        print("OPENROUTER ERROR:", repr(e))
        raise HTTPException(
            status_code=500,
            detail=f"Ошибка OpenRouter: {repr(e)}"
        )