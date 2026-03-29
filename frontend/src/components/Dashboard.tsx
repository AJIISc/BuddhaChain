import React, { useEffect, useState, useCallback } from "react";
import { fetchPending, submitValidation, PendingItem } from "../api";

const VALIDATOR_ID = "validator_001"; // In production, from auth

const Dashboard: React.FC = () => {
  const [items, setItems] = useState<PendingItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState<string | null>(null);
  const [notes, setNotes] = useState<Record<string, string>>({});

  const loadItems = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const data = await fetchPending();
      setItems(data);
    } catch {
      setError("Failed to load pending validations");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadItems();
  }, [loadItems]);

  const handleAction = async (
    dataId: string,
    action: "accept" | "modify" | "reject"
  ) => {
    setSubmitting(dataId);
    try {
      await submitValidation({
        data_id: dataId,
        validator_id: VALIDATOR_ID,
        action,
        notes: notes[dataId] || "",
      });
      setItems((prev) => prev.filter((i) => i.data_id !== dataId));
    } catch {
      setError("Failed to submit validation");
    } finally {
      setSubmitting(null);
    }
  };

  const riskBadge = (risk: string) => {
    const cls =
      risk === "critical"
        ? "badge-critical"
        : risk === "high"
        ? "badge-high"
        : risk === "medium"
        ? "badge-medium"
        : "badge-low";
    return <span className={`badge ${cls}`}>{risk}</span>;
  };

  const confidenceColor = (c: number) =>
    c >= 0.8 ? "var(--success)" : c >= 0.6 ? "var(--warning)" : "var(--danger)";

  if (loading) return <div className="empty-state">Loading...</div>;

  return (
    <div>
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "1rem",
        }}
      >
        <h2>Pending Validations ({items.length})</h2>
        <button className="btn btn-primary" onClick={loadItems}>
          Refresh
        </button>
      </div>

      {error && <div className="message message-error">{error}</div>}

      {items.length === 0 ? (
        <div className="empty-state">
          <p>No items pending validation</p>
        </div>
      ) : (
        items.map((item) => (
          <div className="card" key={item.data_id}>
            <div className="card-header">
              <code style={{ fontSize: "0.8rem", color: "var(--text-muted)" }}>
                {item.data_id}
              </code>
              {riskBadge(String(item.labels?.risk || "low"))}
            </div>

            <div className="labels-grid">
              {Object.entries(item.labels).map(([key, val]) => (
                <React.Fragment key={key}>
                  <span className="key">{key}:</span>
                  <span>{typeof val === "object" ? JSON.stringify(val) : String(val)}</span>
                </React.Fragment>
              ))}
            </div>

            <div style={{ marginTop: "0.75rem" }}>
              <span style={{ fontSize: "0.85rem", color: "var(--text-muted)" }}>
                AI Confidence: {(item.confidence * 100).toFixed(1)}%
              </span>
              <div className="confidence-bar">
                <div
                  className="confidence-fill"
                  style={{
                    width: `${item.confidence * 100}%`,
                    background: confidenceColor(item.confidence),
                  }}
                />
              </div>
            </div>

            <div className="form-group" style={{ marginTop: "0.75rem" }}>
              <label>Notes (optional)</label>
              <input
                type="text"
                placeholder="Add review notes..."
                value={notes[item.data_id] || ""}
                onChange={(e) =>
                  setNotes((prev) => ({ ...prev, [item.data_id]: e.target.value }))
                }
              />
            </div>

            <div className="btn-group">
              <button
                className="btn btn-accept"
                disabled={submitting === item.data_id}
                onClick={() => handleAction(item.data_id, "accept")}
              >
                Accept
              </button>
              <button
                className="btn btn-modify"
                disabled={submitting === item.data_id}
                onClick={() => handleAction(item.data_id, "modify")}
              >
                Modify
              </button>
              <button
                className="btn btn-reject"
                disabled={submitting === item.data_id}
                onClick={() => handleAction(item.data_id, "reject")}
              >
                Reject
              </button>
            </div>
          </div>
        ))
      )}
    </div>
  );
};

export default Dashboard;
