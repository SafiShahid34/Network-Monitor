import { useEffect, useState } from "react";
import "./App.css";

type Device = {
  ip: string;
  mac: string;
  hostname: string;
};

function App() {
  const [devices, setDevices] = useState<Device[]>([]);
  const [loading, setLoading] = useState(false);

  async function fetchDevices() {
    try {
      const res = await fetch("http://localhost:8080/devices");
      const json = await res.json();

      console.log("Fetched from backend:", json);

      // Handles both array and object response formats
      setDevices(Array.isArray(json) ? json : json.devices || []);
    } catch (err) {
      console.error("Error fetching devices:", err);
    }
  }

  async function scanNetwork() {
    setLoading(true);
    try {
      const res = await fetch("http://localhost:8080/scan", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ cidr: "192.168.12.0/24" }),
      });
      console.log("Scan triggered:", await res.json());
      setTimeout(fetchDevices, 3000); // allow backend time to complete scan
    } catch (err) {
      console.error("Scan failed:", err);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    fetchDevices();
  }, []);

  return (
    <div className="dashboard">
      <h1 className="title">🧭 Network Monitor Dashboard</h1>

      <button className="scan-btn" onClick={scanNetwork} disabled={loading}>
        {loading ? "Scanning..." : "Scan Network"}
      </button>

      <div className="table-container">
        <table className="device-table">
          <thead>
            <tr>
              <th>IP Address</th>
              <th>MAC</th>
              <th>Hostname</th>
            </tr>
          </thead>
          <tbody>
            {devices.length === 0 ? (
              <tr>
                <td colSpan={3} className="empty">
                  {loading
                    ? "🔍 Scanning your network..."
                    : "No devices found yet."}
                </td>
              </tr>
            ) : (
              devices.map((d, i) => (
                <tr key={i}>
                  <td>{d.ip}</td>
                  <td>{d.mac || "—"}</td>
                  <td>{d.hostname || "Unknown"}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {devices.length > 0 && !loading && (
        <p className="footer">🌐 {devices.length} devices online</p>
      )}
    </div>
  );
}

export default App;
