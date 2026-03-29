import React, { useState } from "react";
import Dashboard from "./components/Dashboard";
import UploadForm from "./components/UploadForm";
import LabelLookup from "./components/LabelLookup";
import "./App.css";

type Tab = "dashboard" | "upload" | "lookup";

const App: React.FC = () => {
  const [tab, setTab] = useState<Tab>("dashboard");

  return (
    <div className="app">
      <header className="app-header">
        <h1>⛓ SynapseChain</h1>
        <p className="subtitle">Human Validation Dashboard</p>
        <nav className="tabs">
          <button
            className={tab === "dashboard" ? "active" : ""}
            onClick={() => setTab("dashboard")}
          >
            Validation Queue
          </button>
          <button
            className={tab === "upload" ? "active" : ""}
            onClick={() => setTab("upload")}
          >
            Upload Data
          </button>
          <button
            className={tab === "lookup" ? "active" : ""}
            onClick={() => setTab("lookup")}
          >
            Label Lookup
          </button>
        </nav>
      </header>

      <main className="app-main">
        {tab === "dashboard" && <Dashboard />}
        {tab === "upload" && <UploadForm />}
        {tab === "lookup" && <LabelLookup />}
      </main>
    </div>
  );
};

export default App;
