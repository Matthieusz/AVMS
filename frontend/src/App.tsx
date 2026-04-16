import { useState, useEffect } from "react";
import "./App.css";

interface HealthResponse {
  status: string;
  timestamp: string;
  service: string;
}

function App() {
  const [health, setHealth] = useState<HealthResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch("/api/health")
      .then((res) => {
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        return res.json();
      })
      .then((data: HealthResponse) => {
        setHealth(data);
        setError(null);
      })
      .catch((err) => {
        setError(err.message);
      })
      .finally(() => setLoading(false));
  }, []);

  return (
    <section id="center">
      <h1>AVMS</h1>
      <h2>API Connection Status</h2>
      {loading && <p className="status loading">Connecting to API...</p>}
      {error && (
        <div className="status error">
          <p className="badge error-badge">Disconnected</p>
          <p>{error}</p>
        </div>
      )}
      {health && (
        <div className="status connected">
          <p className="badge ok-badge">Connected</p>
          <table>
            <tbody>
              <tr>
                <td className="label">Status</td>
                <td>{health.status}</td>
              </tr>
              <tr>
                <td className="label">Service</td>
                <td>{health.service}</td>
              </tr>
              <tr>
                <td className="label">Timestamp</td>
                <td>{new Date(health.timestamp).toLocaleString()}</td>
              </tr>
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}

export default App;