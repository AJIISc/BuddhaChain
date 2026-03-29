"""SynapseChain AI Labeling Service - Event & Content classification engine."""

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import Optional
import time

from .labeler import EventLabeler, ContentLabeler

app = FastAPI(title="SynapseChain AI Labeling Service", version="0.1.0")

event_labeler = EventLabeler()
content_labeler = ContentLabeler()


class LabelRequest(BaseModel):
    data_id: str
    type: str  # "event" or "content"
    raw_data: dict
    metadata: Optional[dict] = None


class LabelResponse(BaseModel):
    labels: dict
    confidence: float
    model_version: str
    processing_time_ms: int


@app.get("/health")
def health():
    return {"status": "healthy", "service": "ai-labeling"}


@app.post("/label", response_model=LabelResponse)
def label_data(req: LabelRequest):
    start = time.time()

    if req.type == "event":
        result = event_labeler.label(req.raw_data)
    elif req.type == "content":
        result = content_labeler.label(req.raw_data)
    else:
        raise HTTPException(status_code=400, detail="type must be 'event' or 'content'")

    processing_time_ms = int((time.time() - start) * 1000)

    return LabelResponse(
        labels=result["labels"],
        confidence=result["confidence"],
        model_version=result["model_version"],
        processing_time_ms=processing_time_ms,
    )
