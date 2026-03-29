import React, { useState } from "react";
import { uploadData } from "../api";

const UploadForm: React.FC = () => {
  const [type, setType] = useState<"event" | "content">("event");
  const [jsonInput, setJsonInput] = useState(
    JSON.stringify(
      {
        speed: 130,
        deceleration: -0.6,
        angular_velocity: 0.2,
        impact_force: 0.1,
      },
      null,
      2
    )
  );
  const [result, setResult] = useState<{ data_id: string; status: string } | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setResult(null);
    setLoading(true);

    try {
      const data = JSON.parse(jsonInput);
      const res = await uploadData(type, data);
      setResult(res);
    } catch (err: unknown) {
      if (err instanceof SyntaxError) {
        setError("Invalid JSON input");
      } else {
        setError("Upload failed");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <h2 style={{ marginBottom: "1rem" }}>Upload Data</h2>

      {result && (
        <div className="message message-success">
          Uploaded! Data ID: <code>{result.data_id}</code> — Status: {result.status}
        </div>
      )}
      {error && <div className="message message-error">{error}</div>}

      <form onSubmit={handleSubmit} className="card">
        <div className="form-group">
          <label>Data Type</label>
          <select value={type} onChange={(e) => setType(e.target.value as "event" | "content")}>
            <option value="event">Event (Sensor/Driver)</option>
            <option value="content">Content (Text)</option>
          </select>
        </div>

        <div className="form-group">
          <label>Raw Data (JSON)</label>
          <textarea
            rows={10}
            value={jsonInput}
            onChange={(e) => setJsonInput(e.target.value)}
            style={{ fontFamily: "monospace", fontSize: "0.85rem" }}
          />
        </div>

        <button className="btn btn-primary" type="submit" disabled={loading}>
          {loading ? "Uploading..." : "Upload"}
        </button>
      </form>

      <div className="card" style={{ marginTop: "1rem" }}>
        <h3 style={{ marginBottom: "0.5rem", fontSize: "1rem" }}>Example Payloads</h3>
        <p style={{ fontSize: "0.85rem", color: "var(--text-muted)", marginBottom: "0.5rem" }}>
          Click to use:
        </p>
        <div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap" }}>
          <button
            className="btn btn-primary"
            type="button"
            onClick={() => {
              setType("event");
              setJsonInput(
                JSON.stringify(
                  { speed: 130, deceleration: -0.3, angular_velocity: 0.1 },
                  null,
                  2
                )
              );
            }}
          >
            Speeding Event
          </button>
          <button
            className="btn btn-primary"
            type="button"
            onClick={() => {
              setType("event");
              setJsonInput(
                JSON.stringify(
                  { impact_force: 0.9, deceleration: -0.8, speed: 60 },
                  null,
                  2
                )
              );
            }}
          >
            Collision Event
          </button>
          <button
            className="btn btn-primary"
            type="button"
            onClick={() => {
              setType("content");
              setJsonInput(
                JSON.stringify(
                  { text: "Emergency! Critical system failure detected immediately." },
                  null,
                  2
                )
              );
            }}
          >
            Urgent Content
          </button>
        </div>
      </div>
    </div>
  );
};

export default UploadForm;
