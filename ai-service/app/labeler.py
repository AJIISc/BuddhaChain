"""
AI Labeling engines for events and content.

MVP: rule-based classifiers that can be swapped for ML models later.
"""

import numpy as np


class EventLabeler:
    """Classifies sensor/driver event data."""

    MODEL_VERSION = "event-rules-v0.1"

    # Thresholds for event classification
    EVENT_RULES = {
        "hard_braking": {"field": "deceleration", "threshold": -0.5, "risk": "medium"},
        "sudden_acceleration": {"field": "acceleration", "threshold": 0.7, "risk": "low"},
        "sharp_turn": {"field": "angular_velocity", "threshold": 0.6, "risk": "medium"},
        "collision": {"field": "impact_force", "threshold": 0.8, "risk": "critical"},
        "speeding": {"field": "speed", "threshold": 120, "risk": "high"},
        "drowsiness": {"field": "eye_closure_ratio", "threshold": 0.7, "risk": "high"},
    }

    def label(self, raw_data: dict) -> dict:
        detected_events = []
        max_risk = "low"
        risk_order = {"low": 0, "medium": 1, "high": 2, "critical": 3}

        for event_type, rule in self.EVENT_RULES.items():
            field = rule["field"]
            if field in raw_data:
                value = float(raw_data[field])
                threshold = rule["threshold"]

                # For negative thresholds (braking), check if value is below
                if threshold < 0:
                    triggered = value <= threshold
                else:
                    triggered = value >= threshold

                if triggered:
                    detected_events.append(event_type)
                    if risk_order.get(rule["risk"], 0) > risk_order.get(max_risk, 0):
                        max_risk = rule["risk"]

        # If no specific event detected, classify as normal
        if not detected_events:
            event_type = raw_data.get("event_type", "normal_driving")
            detected_events = [event_type]
            max_risk = "low"

        # Confidence: higher for clear-cut events, lower for ambiguous
        confidence = self._compute_confidence(raw_data, detected_events)

        return {
            "labels": {
                "event_type": detected_events[0] if len(detected_events) == 1 else detected_events,
                "risk": max_risk,
                "all_events": detected_events,
            },
            "confidence": confidence,
            "model_version": self.MODEL_VERSION,
        }

    def _compute_confidence(self, raw_data: dict, events: list) -> float:
        """Heuristic confidence based on signal clarity."""
        if not events or events == ["normal_driving"]:
            return 0.65

        # Count how many data fields were present
        available_fields = sum(1 for rule in self.EVENT_RULES.values() if rule["field"] in raw_data)
        total_fields = len(self.EVENT_RULES)

        data_completeness = available_fields / max(total_fields, 1)

        # Base confidence + data completeness bonus
        base = 0.7
        confidence = base + (data_completeness * 0.25)

        # Add some noise to simulate model uncertainty
        noise = np.random.uniform(-0.05, 0.05)
        confidence = np.clip(confidence + noise, 0.1, 0.99)

        return round(float(confidence), 3)


class ContentLabeler:
    """Classifies text/content data."""

    MODEL_VERSION = "content-rules-v0.1"

    CATEGORIES = {
        "positive": ["good", "great", "excellent", "amazing", "love", "best", "happy", "wonderful"],
        "negative": ["bad", "terrible", "awful", "hate", "worst", "poor", "horrible", "angry"],
        "neutral": ["okay", "fine", "average", "normal", "standard"],
        "urgent": ["emergency", "critical", "urgent", "immediate", "danger", "alert", "warning"],
    }

    RISK_MAP = {
        "positive": "low",
        "negative": "medium",
        "neutral": "low",
        "urgent": "high",
    }

    def label(self, raw_data: dict) -> dict:
        text = raw_data.get("text", raw_data.get("content", ""))
        if not isinstance(text, str):
            text = str(text)

        text_lower = text.lower()
        scores = {}

        for category, keywords in self.CATEGORIES.items():
            score = sum(1 for kw in keywords if kw in text_lower)
            scores[category] = score

        # Pick highest scoring category
        if max(scores.values()) == 0:
            category = "neutral"
            confidence = 0.5
        else:
            category = max(scores, key=scores.get)
            total_matches = sum(scores.values())
            confidence = min(0.95, 0.6 + (scores[category] / max(total_matches, 1)) * 0.3)

        risk = self.RISK_MAP.get(category, "low")

        return {
            "labels": {
                "category": category,
                "risk": risk,
                "keyword_scores": scores,
            },
            "confidence": round(confidence, 3),
            "model_version": self.MODEL_VERSION,
        }
