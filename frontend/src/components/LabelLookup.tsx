import React, { useState } from "react";
import { fetchLabel, FinalLabel } from "../api";

const LabelLookup: React.FC = () => {
  const [dataId, setDataId] = useState("");
  const [label, setLabel] = useState<FinalLabel | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleLookup = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!dataId.trim()) return;

    setError("");
    setLabel(null);
    setLoading(true);

    try {
      const res = await fetchLabel(dataId.trim());
      setLabel(res);
    } catch {
      setError("Label not found for this data ID");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <h2 style={{ marginBottom: "1rem" }}>Label Lookup</h2>

      <form onSubmit={handleLookup} className="card">
        <div className="form-group">
          <label>Data ID (UUID)</label>
          <input
            type="text"
            placeholder="e.g. a1b2c3d4-e5f6-7890-abcd-ef1234567890"
            value={dataId}
            onChange={(e) => setDataId(e.target.value)}
          />
        </div>
        <button className="btn btn-primary" type="submit" disabled={loading}>
          {loading ? "Looking up..." : "Lookup"}
        </button>
      </form>

      {error && (
        <div className="message message-error" style={{ marginTop: "1rem" }}>
          {error}
        </div>
      )}

      {label && (
        <div className="card" style={{ marginTop: "1rem" }}>
          <h3 style={{ marginBottom: "0.75rem" }}>Final Label</h3>

          <div className="labels-grid">
            {Object.entries(label.final_label).map(([key, val]) => (
              <React.Fragment key={key}>
                <span className="key">{key}:</span>
                <span>
                  {typeof val === "object" ? JSON.stringify(val) : String(val)}
                </span>
              </React.Fragment>
            ))}
          </div>

          <div style={{ marginTop: "1rem" }}>
            <span style={{ color: "var(--text-muted)", fontSize: "0.85rem" }}>
              Confidence: {(label.confidence * 100).toFixed(1)}%
            </span>
            <div className="confidence-bar">
              <div
                className="confidence-fill"
                style={{
                  width: `${label.confidence * 100}%`,
                  background:
                    label.confidence >= 0.8
                      ? "var(--success)"
                      : label.confidence >= 0.6
                      ? "var(--warning)"
                      : "var(--danger)",
                }}
              />
            </div>
          </div>

          <div style={{ marginTop: "0.75rem" }}>
            <span style={{ color: "var(--text-muted)", fontSize: "0.85rem" }}>
              Sources:{" "}
            </span>
            {label.source.map((s) => (
              <span key={s} className="badge badge-low" style={{ marginRight: "0.3rem" }}>
                {s}
              </span>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

export default LabelLookup;
