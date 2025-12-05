package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// DemoPage serves the interactive demo and documentation page
func DemoPage(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(demoPageHTML)
}

// demoPageHTML contains the complete HTML for the interactive demo page
var demoPageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go Fiber Backend API - Demo & Documentation</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            line-height: 1.6;
            color: #333;
            background: #f5f5f5;
        }
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 2rem 0;
            margin-bottom: 2rem;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        header h1 { font-size: 2.5rem; margin-bottom: 0.5rem; }
        header p { font-size: 1.1rem; opacity: 0.9; }
        .nav {
            background: white;
            padding: 1rem;
            border-radius: 8px;
            margin-bottom: 2rem;
            box-shadow: 0 2px 5px rgba(0,0,0,0.1);
            position: sticky;
            top: 10px;
            z-index: 100;
        }
        .nav-links {
            display: flex;
            flex-wrap: wrap;
            gap: 1rem;
            list-style: none;
        }
        .nav-links a {
            color: #667eea;
            text-decoration: none;
            padding: 0.5rem 1rem;
            border-radius: 4px;
            transition: background 0.3s;
        }
        .nav-links a:hover { background: #f0f0f0; }
        .section {
            background: white;
            padding: 2rem;
            margin-bottom: 2rem;
            border-radius: 8px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.1);
        }
        .section h2 {
            color: #667eea;
            margin-bottom: 1rem;
            padding-bottom: 0.5rem;
            border-bottom: 2px solid #667eea;
        }
        .section h3 {
            color: #764ba2;
            margin-top: 1.5rem;
            margin-bottom: 0.5rem;
        }
        .code-block {
            background: #282c34;
            color: #abb2bf;
            padding: 1rem;
            border-radius: 4px;
            overflow-x: auto;
            margin: 1rem 0;
            font-family: 'Courier New', monospace;
            font-size: 0.9rem;
        }
        .code-block code { color: #abb2bf; }
        .btn {
            background: #667eea;
            color: white;
            border: none;
            padding: 0.75rem 1.5rem;
            border-radius: 4px;
            cursor: pointer;
            font-size: 1rem;
            transition: background 0.3s;
            margin: 0.5rem 0.5rem 0.5rem 0;
        }
        .btn:hover { background: #5568d3; }
        .btn-success { background: #28a745; }
        .btn-success:hover { background: #218838; }
        .btn-danger { background: #dc3545; }
        .btn-danger:hover { background: #c82333; }
        .input-group {
            margin: 1rem 0;
        }
        .input-group label {
            display: block;
            margin-bottom: 0.5rem;
            font-weight: 600;
            color: #555;
        }
        .input-group input,
        .input-group textarea,
        .input-group select {
            width: 100%;
            padding: 0.75rem;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 1rem;
        }
        .input-group textarea {
            min-height: 150px;
            font-family: 'Courier New', monospace;
        }
        .response-box {
            background: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 4px;
            padding: 1rem;
            margin-top: 1rem;
            max-height: 400px;
            overflow-y: auto;
        }
        .response-box pre {
            margin: 0;
            white-space: pre-wrap;
            word-wrap: break-word;
        }
        .status-badge {
            display: inline-block;
            padding: 0.25rem 0.75rem;
            border-radius: 12px;
            font-size: 0.85rem;
            font-weight: 600;
            margin-left: 0.5rem;
        }
        .status-connected { background: #d4edda; color: #155724; }
        .status-disconnected { background: #f8d7da; color: #721c24; }
        .ws-messages {
            background: #282c34;
            color: #abb2bf;
            padding: 1rem;
            border-radius: 4px;
            max-height: 300px;
            overflow-y: auto;
            font-family: 'Courier New', monospace;
            font-size: 0.85rem;
            margin-top: 1rem;
        }
        .ws-message {
            padding: 0.5rem;
            border-bottom: 1px solid #3e4451;
        }
        .ws-message:last-child { border-bottom: none; }
        .tabs {
            display: flex;
            border-bottom: 2px solid #dee2e6;
            margin-bottom: 1rem;
        }
        .tab {
            padding: 1rem 1.5rem;
            cursor: pointer;
            border: none;
            background: none;
            font-size: 1rem;
            color: #666;
            transition: all 0.3s;
        }
        .tab.active {
            color: #667eea;
            border-bottom: 2px solid #667eea;
            margin-bottom: -2px;
        }
        .tab-content {
            display: none;
        }
        .tab-content.active {
            display: block;
        }
        .alert {
            padding: 1rem;
            border-radius: 4px;
            margin: 1rem 0;
        }
        .alert-info {
            background: #d1ecf1;
            color: #0c5460;
            border: 1px solid #bee5eb;
        }
        .alert-success {
            background: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }
        .alert-warning {
            background: #fff3cd;
            color: #856404;
            border: 1px solid #ffeaa7;
        }
        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 1rem;
            margin: 1rem 0;
        }
        .card {
            background: #f8f9fa;
            padding: 1.5rem;
            border-radius: 4px;
            border: 1px solid #dee2e6;
        }
        .card h4 {
            color: #667eea;
            margin-bottom: 0.5rem;
        }
        .copy-btn {
            background: #6c757d;
            color: white;
            border: none;
            padding: 0.5rem 1rem;
            border-radius: 4px;
            cursor: pointer;
            font-size: 0.85rem;
            margin-left: 0.5rem;
        }
        .copy-btn:hover { background: #5a6268; }
        @media (max-width: 768px) {
            .nav-links { flex-direction: column; }
            header h1 { font-size: 1.8rem; }
        }
    </style>
</head>
<body>
    <header>
        <div class="container">
            <h1>🚀 Go Fiber Backend API</h1>
            <p>Interactive Demo & Complete Documentation</p>
        </div>
    </header>

    <div class="container">
        <nav class="nav">
            <ul class="nav-links">
                <li><a href="#getting-started">Getting Started</a></li>
                <li><a href="#api-tester">API Tester</a></li>
                <li><a href="#websocket">WebSocket</a></li>
                <li><a href="#graphql">GraphQL</a></li>
                <li><a href="#auth">Authentication</a></li>
                <li><a href="#rate-limit">Rate Limiter</a></li>
                <li><a href="#integrations">Frontend Integration</a></li>
                <li><a href="#deployment">Deployment</a></li>
            </ul>
        </nav>

        <!-- Getting Started -->
        <section id="getting-started" class="section">
            <h2>📚 Getting Started</h2>
            <div class="alert alert-info">
                <strong>Welcome!</strong> This is a standalone Go Fiber backend API. Your frontend apps connect to it via HTTP and WebSocket.
            </div>
            
            <h3>Quick Setup Steps</h3>
            <ol style="margin-left: 2rem; line-height: 2;">
                <li>Clone the repository: <code>git clone &lt;repo-url&gt;</code></li>
                <li>Install dependencies: <code>go mod download</code></li>
                <li>Create <code>.env</code> file from <code>.env.example</code></li>
                <li>Configure Supabase URL and keys</li>
                <li>Run: <code>go run ./cmd/server</code></li>
                <li>Visit this page at <code>http://localhost:3000/demo</code></li>
            </ol>

            <h3>API Base URL</h3>
            <div class="code-block">
                <code id="baseUrl">http://localhost:3000</code>
                <button class="copy-btn" onclick="copyToClipboard('baseUrl')">Copy</button>
            </div>
        </section>

        <!-- API Tester -->
        <section id="api-tester" class="section">
            <h2>🧪 API Endpoint Tester</h2>
            <p>Test any API endpoint directly from this page.</p>
            
            <div class="input-group">
                <label>HTTP Method</label>
                <select id="httpMethod">
                    <option>GET</option>
                    <option>POST</option>
                    <option>PUT</option>
                    <option>DELETE</option>
                </select>
            </div>

            <div class="input-group">
                <label>Endpoint URL</label>
                <input type="text" id="endpointUrl" value="/health" placeholder="/api/profile">
            </div>

            <div class="input-group">
                <label>Authorization Token (Bearer)</label>
                <input type="text" id="authToken" placeholder="Leave empty for public endpoints">
            </div>

            <div class="input-group">
                <label>Request Body (JSON)</label>
                <textarea id="requestBody" placeholder='{"key": "value"}'></textarea>
            </div>

            <button class="btn btn-success" onclick="testEndpoint()">Send Request</button>
            <button class="btn" onclick="clearResponse()">Clear</button>

            <div id="apiResponse" class="response-box" style="display: none;">
                <h4>Response:</h4>
                <pre id="responseContent"></pre>
            </div>
        </section>

        <!-- WebSocket Tester -->
        <section id="websocket" class="section">
            <h2>🔌 WebSocket Connection Tester</h2>
            <p>Connect to the WebSocket endpoint and receive real-time messages.</p>
            
            <div>
                <span>Connection Status:</span>
                <span id="wsStatus" class="status-badge status-disconnected">Disconnected</span>
            </div>

            <div style="margin: 1rem 0;">
                <button class="btn btn-success" id="wsConnectBtn" onclick="connectWebSocket()">Connect</button>
                <button class="btn btn-danger" id="wsDisconnectBtn" onclick="disconnectWebSocket()" disabled>Disconnect</button>
                <button class="btn" onclick="clearWebSocketMessages()">Clear Messages</button>
            </div>

            <div class="input-group">
                <label>Send Message (JSON)</label>
                <textarea id="wsMessage" placeholder='{"message": "test"}'></textarea>
                <button class="btn" onclick="sendWebSocketMessage()" style="margin-top: 0.5rem;">Send Message</button>
            </div>

            <div id="wsMessages" class="ws-messages" style="display: none;">
                <div style="color: #61afef; margin-bottom: 0.5rem;">WebSocket Messages:</div>
                <div id="wsMessagesContent"></div>
            </div>
        </section>

        <!-- GraphQL Tester -->
        <section id="graphql" class="section">
            <h2>🔷 GraphQL Query Tester</h2>
            <p>Test GraphQL queries against the Supabase GraphQL proxy.</p>
            
            <div class="input-group">
                <label>GraphQL Query</label>
                <textarea id="graphqlQuery" placeholder='query { artists { id name } }'>query {
  artists {
    id
    name
    currentPrice
  }
}</textarea>
            </div>

            <div class="input-group">
                <label>Variables (JSON)</label>
                <textarea id="graphqlVariables" placeholder='{}'></textarea>
            </div>

            <div class="input-group">
                <label>Authorization Token (optional)</label>
                <input type="text" id="graphqlToken" placeholder="Bearer token">
            </div>

            <button class="btn btn-success" onclick="testGraphQL()">Execute Query</button>

            <div id="graphqlResponse" class="response-box" style="display: none;">
                <h4>Response:</h4>
                <pre id="graphqlResponseContent"></pre>
            </div>
        </section>

        <!-- Authentication -->
        <section id="auth" class="section">
            <h2>🔐 Authentication Guide</h2>
            
            <h3>How Authentication Works</h3>
            <ol style="margin-left: 2rem; line-height: 2;">
                <li>Obtain JWT token from Supabase Auth (or your auth provider)</li>
                <li>Include token in request header: <code>Authorization: Bearer &lt;token&gt;</code></li>
                <li>Backend validates token and extracts user ID</li>
                <li>Protected routes require valid token</li>
            </ol>

            <h3>JWT Token Decoder</h3>
            <div class="input-group">
                <label>Paste JWT Token</label>
                <input type="text" id="jwtToken" placeholder="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...">
                <button class="btn" onclick="decodeJWT()" style="margin-top: 0.5rem;">Decode Token</button>
            </div>

            <div id="jwtDecoded" class="response-box" style="display: none;">
                <h4>Decoded Token:</h4>
                <pre id="jwtContent"></pre>
            </div>

            <h3>Test Protected Endpoint</h3>
            <p>Use the API Tester above with an Authorization token to test protected endpoints like <code>/api/profile</code>.</p>
        </section>

        <!-- Rate Limiter Tester -->
        <section id="rate-limit" class="section">
            <h2>⚡ Rate Limiter Tester</h2>
            <p>Test the rate limiting functionality by sending multiple requests quickly. Watch the graph to see when rate limits are hit.</p>
            
            <div class="alert alert-info">
                <strong>Note:</strong> Default rate limit is 100 requests per minute. Protected endpoints under <code>/api/*</code> have rate limiting enabled.
            </div>

            <div class="grid" style="grid-template-columns: 1fr 1fr;">
                <div class="card">
                    <h4>Test Configuration</h4>
                    <div class="input-group">
                        <label>Endpoint to Test</label>
                        <select id="rateLimitEndpoint">
                            <option value="/health">/health (Public)</option>
                            <option value="/api/profile">/api/profile (Protected - requires auth)</option>
                        </select>
                    </div>
                    <div class="input-group">
                        <label>Number of Requests</label>
                        <input type="number" id="rateLimitCount" value="50" min="1" max="200">
                    </div>
                    <div class="input-group">
                        <label>Delay Between Requests (ms)</label>
                        <input type="number" id="rateLimitDelay" value="100" min="10" max="5000">
                    </div>
                    <div class="input-group">
                        <label>Authorization Token (optional, for protected endpoints)</label>
                        <input type="text" id="rateLimitToken" placeholder="Bearer token">
                    </div>
                    <button class="btn btn-success" onclick="startRateLimitTest()">Start Test</button>
                    <button class="btn btn-danger" onclick="stopRateLimitTest()">Stop Test</button>
                    <button class="btn" onclick="resetRateLimitTest()">Reset</button>
                </div>
                <div class="card">
                    <h4>Statistics</h4>
                    <div style="font-size: 1.1rem; line-height: 2;">
                        <div>Total Requests: <strong id="rateLimitTotal">0</strong></div>
                        <div>Successful: <strong id="rateLimitSuccess" style="color: #28a745;">0</strong></div>
                        <div>Rate Limited: <strong id="rateLimitLimited" style="color: #dc3545;">0</strong></div>
                        <div>Errors: <strong id="rateLimitErrors" style="color: #ffc107;">0</strong></div>
                        <div>Success Rate: <strong id="rateLimitSuccessRate">0%</strong></div>
                    </div>
                </div>
            </div>

            <div class="input-group" style="margin-top: 1.5rem;">
                <label>Request Timeline Graph</label>
                <div style="background: #f8f9fa; border: 1px solid #dee2e6; border-radius: 4px; padding: 1rem; min-height: 300px; overflow-x: auto;">
                    <canvas id="rateLimitChart" style="max-width: 100%; height: auto;"></canvas>
                </div>
            </div>

            <div class="input-group">
                <label>Request Log</label>
                <div id="rateLimitLog" class="ws-messages" style="max-height: 200px; display: block;">
                    <div style="color: #61afef; margin-bottom: 0.5rem;">Request History:</div>
                    <div id="rateLimitLogContent"></div>
                </div>
            </div>
        </section>

        <!-- Frontend Integration -->
        <section id="integrations" class="section">
            <h2>💻 Frontend Integration Examples</h2>
            <p>Complete code examples for connecting your frontend apps to this backend.</p>

            <div class="tabs">
                <button class="tab active" onclick="showTab('react')">React</button>
                <button class="tab" onclick="showTab('nextjs')">Next.js</button>
                <button class="tab" onclick="showTab('flutter')">Flutter</button>
                <button class="tab" onclick="showTab('reactnative')">React Native</button>
            </div>

            <div id="react" class="tab-content active">
                <h3>React Integration</h3>
                <div class="code-block">
                    <code id="reactCode">// API Configuration
const API_URL = 'http://localhost:3000';

// Making authenticated requests
async function fetchProfile(token) {
  const response = await fetch(API_URL + '/api/profile', {
    headers: {
      'Authorization': 'Bearer ' + token,
      'Content-Type': 'application/json'
    }
  });
  return response.json();
}

// WebSocket connection
const ws = new WebSocket('ws://localhost:3000/ws');
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Received:', data);
};

// Using in React component
function ProfileComponent() {
  const [profile, setProfile] = useState(null);
  const token = localStorage.getItem('token');

  useEffect(() => {
    fetchProfile(token).then(setProfile);
  }, [token]);

  return <div>{profile && <p>User: {profile.user}</p>}</div>;
}</code>
                    <button class="copy-btn" onclick="copyToClipboard('reactCode')">Copy</button>
                </div>
            </div>

            <div id="nextjs" class="tab-content">
                <h3>Next.js Integration</h3>
                <div class="code-block">
                    <code id="nextjsCode">// Server Component (app directory)
async function getProfile(token) {
  const res = await fetch(process.env.BACKEND_URL + '/api/profile', {
    headers: { 'Authorization': 'Bearer ' + token },
    cache: 'no-store'
  });
  return res.json();
}

// Client Component
'use client';
import { useEffect, useState } from 'react';

export default function Profile() {
  const [data, setData] = useState(null);
  const token = localStorage.getItem('token');

  useEffect(() => {
    fetch(process.env.NEXT_PUBLIC_BACKEND_URL + '/api/profile', {
      headers: { 'Authorization': 'Bearer ' + token }
    })
    .then(res => res.json())
    .then(setData);
  }, [token]);

  return <div>{data?.user}</div>;
}</code>
                    <button class="copy-btn" onclick="copyToClipboard('nextjsCode')">Copy</button>
                </div>
            </div>

            <div id="flutter" class="tab-content">
                <h3>Flutter Integration</h3>
                <div class="code-block">
                    <code id="flutterCode">// pubspec.yaml
dependencies:
  http: ^1.1.0
  web_socket_channel: ^2.4.0

// API Service
import 'package:http/http.dart' as http;

class ApiService {
  static const String baseUrl = 'http://localhost:3000';

  Future<Map<String, dynamic>> getProfile(String token) async {
    final response = await http.get(
      Uri.parse('$baseUrl/api/profile'),
      headers: {
        'Authorization': 'Bearer $token',
        'Content-Type': 'application/json',
      },
    );
    return json.decode(response.body);
  }
}

// WebSocket
import 'package:web_socket_channel/web_socket_channel.dart';

final channel = WebSocketChannel.connect(
  Uri.parse('ws://localhost:3000/ws'),
);

channel.stream.listen((message) {
  final data = json.decode(message);
  print('Received: $data');
});</code>
                    <button class="copy-btn" onclick="copyToClipboard('flutterCode')">Copy</button>
                </div>
            </div>

            <div id="reactnative" class="tab-content">
                <h3>React Native Integration</h3>
                <div class="code-block">
                    <code id="reactnativeCode">// Install: npm install @react-native-async-storage/async-storage

import AsyncStorage from '@react-native-async-storage/async-storage';

const API_URL = 'http://localhost:3000';

// Store token
await AsyncStorage.setItem('token', userToken);

// Fetch with token
const token = await AsyncStorage.getItem('token');
const response = await fetch(API_URL + '/api/profile', {
  headers: {
        'Authorization': 'Bearer ' + token,
    'Content-Type': 'application/json'
  }
});

// WebSocket
const ws = new WebSocket('ws://localhost:3000/ws');
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Received:', data);
};</code>
                    <button class="copy-btn" onclick="copyToClipboard('reactnativeCode')">Copy</button>
                </div>
            </div>
        </section>

        <!-- Deployment -->
        <section id="deployment" class="section">
            <h2>🚀 Deployment Guides</h2>
            
            <div class="grid">
                <div class="card">
                    <h4>Fly.io</h4>
                    <ol style="margin-left: 1.5rem; line-height: 1.8;">
                        <li>Install Fly CLI</li>
                        <li>Run: <code>fly launch</code></li>
                        <li>Set secrets: <code>fly secrets set KEY=value</code></li>
                        <li>Deploy: <code>fly deploy</code></li>
                    </ol>
                </div>
                <div class="card">
                    <h4>Render</h4>
                    <ol style="margin-left: 1.5rem; line-height: 1.8;">
                        <li>Connect GitHub repo</li>
                        <li>Create Web Service</li>
                        <li>Configure environment variables</li>
                        <li>Auto-deploy on push</li>
                    </ol>
                </div>
                <div class="card">
                    <h4>Railway</h4>
                    <ol style="margin-left: 1.5rem; line-height: 1.8;">
                        <li>Install Railway CLI</li>
                        <li>Run: <code>railway init</code></li>
                        <li>Set variables: <code>railway variables set KEY=value</code></li>
                        <li>Deploy: <code>railway up</code></li>
                    </ol>
                </div>
            </div>

            <div class="alert alert-warning">
                <strong>Important:</strong> Set <code>ALLOWED_ORIGINS</code> to your frontend domain(s) in production!
            </div>
        </section>
    </div>

    <script>
        // Update base URL
        const baseUrl = window.location.origin;
        document.getElementById('baseUrl').textContent = baseUrl;

        let ws = null;

        // API Tester
        async function testEndpoint() {
            const method = document.getElementById('httpMethod').value;
            const url = document.getElementById('endpointUrl').value;
            const token = document.getElementById('authToken').value;
            const body = document.getElementById('requestBody').value;

            const headers = {
                'Content-Type': 'application/json'
            };
            if (token) {
                headers['Authorization'] = 'Bearer ' + token;
            }

            const options = {
                method: method,
                headers: headers
            };

            if (body && (method === 'POST' || method === 'PUT')) {
                options.body = body;
            }

            try {
                const response = await fetch(baseUrl + url, options);
                const data = await response.json();
                document.getElementById('responseContent').textContent = 
                    JSON.stringify(data, null, 2);
                document.getElementById('apiResponse').style.display = 'block';
            } catch (error) {
                document.getElementById('responseContent').textContent = 
                    'Error: ' + error.message;
                document.getElementById('apiResponse').style.display = 'block';
            }
        }

        function clearResponse() {
            document.getElementById('apiResponse').style.display = 'none';
            document.getElementById('responseContent').textContent = '';
        }

        // WebSocket
        function connectWebSocket() {
            const wsUrl = baseUrl.replace('http://', 'ws://').replace('https://', 'wss://') + '/ws';
            ws = new WebSocket(wsUrl);

            ws.onopen = () => {
                document.getElementById('wsStatus').textContent = 'Connected';
                document.getElementById('wsStatus').className = 'status-badge status-connected';
                document.getElementById('wsConnectBtn').disabled = true;
                document.getElementById('wsDisconnectBtn').disabled = false;
                document.getElementById('wsMessages').style.display = 'block';
                addWsMessage('Connected to WebSocket', 'system');
            };

            ws.onmessage = (event) => {
                addWsMessage(event.data, 'received');
            };

            ws.onerror = (error) => {
                addWsMessage('Error: ' + error, 'error');
            };

            ws.onclose = () => {
                document.getElementById('wsStatus').textContent = 'Disconnected';
                document.getElementById('wsStatus').className = 'status-badge status-disconnected';
                document.getElementById('wsConnectBtn').disabled = false;
                document.getElementById('wsDisconnectBtn').disabled = true;
                addWsMessage('Disconnected from WebSocket', 'system');
            };
        }

        function disconnectWebSocket() {
            if (ws) {
                ws.close();
                ws = null;
            }
        }

        function sendWebSocketMessage() {
            const message = document.getElementById('wsMessage').value;
            if (ws && ws.readyState === WebSocket.OPEN) {
                ws.send(message);
                addWsMessage('Sent: ' + message, 'sent');
                document.getElementById('wsMessage').value = '';
            } else {
                alert('WebSocket not connected');
            }
        }

        function addWsMessage(message, type) {
            const content = document.getElementById('wsMessagesContent');
            const div = document.createElement('div');
            div.className = 'ws-message';
            const color = type === 'sent' ? '#98c379' : type === 'error' ? '#e06c75' : '#61afef';
            div.innerHTML = '<span style="color: ' + color + '">[' + new Date().toLocaleTimeString() + ']</span> ' + message;
            content.appendChild(div);
            content.scrollTop = content.scrollHeight;
        }

        function clearWebSocketMessages() {
            document.getElementById('wsMessagesContent').innerHTML = '';
        }

        // GraphQL
        async function testGraphQL() {
            const query = document.getElementById('graphqlQuery').value;
            const variables = document.getElementById('graphqlVariables').value;
            const token = document.getElementById('graphqlToken').value;

            const body = {
                query: query
            };

            if (variables) {
                try {
                    body.variables = JSON.parse(variables);
                } catch (e) {
                    alert('Invalid JSON in variables');
                    return;
                }
            }

            const headers = {
                'Content-Type': 'application/json'
            };
            if (token) {
                headers['Authorization'] = 'Bearer ' + token;
            }

            try {
                const response = await fetch(baseUrl + '/graphql', {
                    method: 'POST',
                    headers: headers,
                    body: JSON.stringify(body)
                });
                const data = await response.json();
                document.getElementById('graphqlResponseContent').textContent = 
                    JSON.stringify(data, null, 2);
                document.getElementById('graphqlResponse').style.display = 'block';
            } catch (error) {
                document.getElementById('graphqlResponseContent').textContent = 
                    'Error: ' + error.message;
                document.getElementById('graphqlResponse').style.display = 'block';
            }
        }

        // JWT Decoder
        function decodeJWT() {
            const token = document.getElementById('jwtToken').value;
            if (!token) {
                alert('Please enter a JWT token');
                return;
            }

            try {
                const parts = token.split('.');
                if (parts.length !== 3) {
                    throw new Error('Invalid JWT format');
                }

                const header = JSON.parse(atob(parts[0].replace(/-/g, '+').replace(/_/g, '/')));
                const payload = JSON.parse(atob(parts[1].replace(/-/g, '+').replace(/_/g, '/')));

                document.getElementById('jwtContent').textContent = 
                    JSON.stringify({ header, payload }, null, 2);
                document.getElementById('jwtDecoded').style.display = 'block';
            } catch (error) {
                document.getElementById('jwtContent').textContent = 
                    'Error decoding token: ' + error.message;
                document.getElementById('jwtDecoded').style.display = 'block';
            }
        }

        // Tabs
        function showTab(tabName) {
            document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
            event.target.classList.add('active');
            document.getElementById(tabName).classList.add('active');
        }

        // Rate Limiter Test
        let rateLimitTestRunning = false;
        let rateLimitTestData = [];
        let rateLimitStats = { total: 0, success: 0, limited: 0, errors: 0 };

        function startRateLimitTest() {
            if (rateLimitTestRunning) {
                alert('Test already running');
                return;
            }

            const endpoint = document.getElementById('rateLimitEndpoint').value;
            const count = parseInt(document.getElementById('rateLimitCount').value);
            const delay = parseInt(document.getElementById('rateLimitDelay').value);
            const token = document.getElementById('rateLimitToken').value;

            if (count < 1 || count > 200) {
                alert('Request count must be between 1 and 200');
                return;
            }

            rateLimitTestRunning = true;
            rateLimitStats = { total: 0, success: 0, limited: 0, errors: 0 };
            rateLimitTestData = [];
            updateRateLimitStats();
            clearRateLimitLog();

            const headers = {
                'Content-Type': 'application/json'
            };
            if (token) {
                headers['Authorization'] = 'Bearer ' + token;
            }

            let requestIndex = 0;
            const startTime = Date.now();

            async function sendRequest() {
                if (!rateLimitTestRunning || requestIndex >= count) {
                    rateLimitTestRunning = false;
                    return;
                }

                const requestTime = Date.now() - startTime;
                const requestStart = performance.now();

                try {
                    const response = await fetch(baseUrl + endpoint, {
                        method: 'GET',
                        headers: headers
                    });

                    const requestDuration = performance.now() - requestStart;
                    const status = response.status;
                    const isSuccess = status === 200 || status === 201;
                    const isRateLimited = status === 429;

                    rateLimitStats.total++;
                    if (isSuccess) {
                        rateLimitStats.success++;
                    } else if (isRateLimited) {
                        rateLimitStats.limited++;
                    } else {
                        rateLimitStats.errors++;
                    }

                    rateLimitTestData.push({
                        time: requestTime,
                        status: status,
                        duration: requestDuration,
                        isSuccess: isSuccess,
                        isRateLimited: isRateLimited
                    });

                    updateRateLimitStats();
                    drawRateLimitChart();
                    addRateLimitLog(requestIndex + 1, status, requestDuration, isRateLimited);

                } catch (error) {
                    rateLimitStats.total++;
                    rateLimitStats.errors++;
                    rateLimitTestData.push({
                        time: requestTime,
                        status: 0,
                        duration: 0,
                        isSuccess: false,
                        isRateLimited: false
                    });
                    updateRateLimitStats();
                    drawRateLimitChart();
                    addRateLimitLog(requestIndex + 1, 'Error', 0, false);
                }

                requestIndex++;
                if (requestIndex < count && rateLimitTestRunning) {
                    setTimeout(sendRequest, delay);
                } else {
                    rateLimitTestRunning = false;
                }
            }

            sendRequest();
        }

        function stopRateLimitTest() {
            rateLimitTestRunning = false;
        }

        function resetRateLimitTest() {
            stopRateLimitTest();
            rateLimitStats = { total: 0, success: 0, limited: 0, errors: 0 };
            rateLimitTestData = [];
            updateRateLimitStats();
            clearRateLimitLog();
            drawRateLimitChart();
        }

        function updateRateLimitStats() {
            document.getElementById('rateLimitTotal').textContent = rateLimitStats.total;
            document.getElementById('rateLimitSuccess').textContent = rateLimitStats.success;
            document.getElementById('rateLimitLimited').textContent = rateLimitStats.limited;
            document.getElementById('rateLimitErrors').textContent = rateLimitStats.errors;
            
            const successRate = rateLimitStats.total > 0 
                ? Math.round((rateLimitStats.success / rateLimitStats.total) * 100) 
                : 0;
            document.getElementById('rateLimitSuccessRate').textContent = successRate + '%';
        }

        function drawRateLimitChart() {
            const canvas = document.getElementById('rateLimitChart');
            const ctx = canvas.getContext('2d');
            
            // Set canvas size based on container
            const container = canvas.parentElement;
            const width = Math.max(800, container.clientWidth - 40);
            const height = 250;
            canvas.width = width;
            canvas.height = height;

            // Clear canvas
            ctx.clearRect(0, 0, width, height);

            if (rateLimitTestData.length === 0) {
                ctx.fillStyle = '#666';
                ctx.font = '16px sans-serif';
                ctx.textAlign = 'center';
                ctx.fillText('No data yet. Start a test to see the graph.', width / 2, height / 2);
                return;
            }

            // Find max time and duration for scaling
            const maxTime = Math.max(...rateLimitTestData.map(d => d.time), 1000);
            const maxDuration = Math.max(...rateLimitTestData.map(d => d.duration), 100);

            // Draw grid
            ctx.strokeStyle = '#e0e0e0';
            ctx.lineWidth = 1;
            for (let i = 0; i <= 10; i++) {
                const y = (height / 10) * i;
                ctx.beginPath();
                ctx.moveTo(0, y);
                ctx.lineTo(width, y);
                ctx.stroke();
            }

            // Draw axes
            ctx.strokeStyle = '#333';
            ctx.lineWidth = 2;
            ctx.beginPath();
            ctx.moveTo(50, height - 30);
            ctx.lineTo(width - 20, height - 30);
            ctx.lineTo(width - 20, 20);
            ctx.stroke();

            // Draw labels
            ctx.fillStyle = '#333';
            ctx.font = '12px sans-serif';
            ctx.textAlign = 'center';
            ctx.fillText('Time (ms)', width / 2, height - 5);
            ctx.save();
            ctx.translate(15, height / 2);
            ctx.rotate(-Math.PI / 2);
            ctx.fillText('Response Time (ms)', 0, 0);
            ctx.restore();

            // Draw data points
            const pointWidth = (width - 70) / Math.max(rateLimitTestData.length, 1);
            rateLimitTestData.forEach((data, index) => {
                const x = 50 + (index * pointWidth);
                const y = height - 30 - ((data.duration / maxDuration) * (height - 50));
                
                // Color based on status
                if (data.isRateLimited) {
                    ctx.fillStyle = '#dc3545'; // Red for rate limited
                } else if (data.isSuccess) {
                    ctx.fillStyle = '#28a745'; // Green for success
                } else {
                    ctx.fillStyle = '#ffc107'; // Yellow for errors
                }

                // Draw point
                ctx.beginPath();
                ctx.arc(x, y, 4, 0, Math.PI * 2);
                ctx.fill();

                // Draw line to next point
                if (index < rateLimitTestData.length - 1) {
                    const nextX = 50 + ((index + 1) * pointWidth);
                    const nextY = height - 30 - ((rateLimitTestData[index + 1].duration / maxDuration) * (height - 50));
                    ctx.strokeStyle = data.isRateLimited ? '#dc3545' : (data.isSuccess ? '#28a745' : '#ffc107');
                    ctx.lineWidth = 2;
                    ctx.beginPath();
                    ctx.moveTo(x, y);
                    ctx.lineTo(nextX, nextY);
                    ctx.stroke();
                }
            });

            // Draw legend
            ctx.fillStyle = '#333';
            ctx.font = '12px sans-serif';
            ctx.textAlign = 'left';
            const legendY = 30;
            ctx.fillStyle = '#28a745';
            ctx.fillRect(width - 150, legendY, 15, 15);
            ctx.fillStyle = '#333';
            ctx.fillText('Success', width - 130, legendY + 12);
            
            ctx.fillStyle = '#dc3545';
            ctx.fillRect(width - 150, legendY + 20, 15, 15);
            ctx.fillStyle = '#333';
            ctx.fillText('Rate Limited', width - 130, legendY + 32);
            
            ctx.fillStyle = '#ffc107';
            ctx.fillRect(width - 150, legendY + 40, 15, 15);
            ctx.fillStyle = '#333';
            ctx.fillText('Error', width - 130, legendY + 52);
        }

        function addRateLimitLog(index, status, duration, isRateLimited) {
            const logContent = document.getElementById('rateLimitLogContent');
            const div = document.createElement('div');
            div.className = 'ws-message';
            let color = '#98c379';
            let statusText = status;
            if (isRateLimited) {
                color = '#e06c75';
                statusText = '429 Rate Limited';
            } else if (status !== 200 && status !== 201) {
                color = '#e5c07b';
            }
            div.innerHTML = '<span style="color: ' + color + '">[' + new Date().toLocaleTimeString() + ']</span> ' +
                'Request #' + index + ' - Status: ' + statusText + 
                (duration > 0 ? ' - Duration: ' + Math.round(duration) + 'ms' : '');
            logContent.appendChild(div);
            logContent.scrollTop = logContent.scrollHeight;
        }

        function clearRateLimitLog() {
            document.getElementById('rateLimitLogContent').innerHTML = '';
        }

        // Copy to clipboard
        function copyToClipboard(elementId) {
            const element = document.getElementById(elementId);
            const text = element.textContent || element.innerText;
            navigator.clipboard.writeText(text).then(() => {
                alert('Copied to clipboard!');
            });
        }

        // Smooth scroll
        document.querySelectorAll('a[href^="#"]').forEach(anchor => {
            anchor.addEventListener('click', function (e) {
                e.preventDefault();
                const target = document.querySelector(this.getAttribute('href'));
                if (target) {
                    target.scrollIntoView({ behavior: 'smooth', block: 'start' });
                }
            });
        });
    </script>
</body>
</html>`

