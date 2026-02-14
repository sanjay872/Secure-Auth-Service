import { useState } from "react";
import {
  signInWithEmailAndPassword,
  createUserWithEmailAndPassword,
} from "firebase/auth";
import { auth } from "./firebase";
import api, { setAccessToken } from "./api/axios";

export default function App() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [status, setStatus] = useState("");
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [profile, setProfile] = useState<any>(null);

  async function handleSignUp() {
    setStatus("Signing up...");
    try {
      await createUserWithEmailAndPassword(auth, email, password);
      setStatus("Signed up successfully ✅");
    } catch (e: any) {
      setStatus(e.message);
    }
  }

  async function handleSignIn() {
    setStatus("Signing in...");
    try {
      const cred = await signInWithEmailAndPassword(auth, email, password);
      const idToken = await cred.user.getIdToken();

      const res = await api.post("/auth/exchange", {
        id_token: idToken,
      });

      setAccessToken(res.data.access_token);

      const profileRes = await api.get("/profile");
      setProfile(profileRes.data);

      setIsAuthenticated(true);
      setStatus("");

    } catch (e: any) {
      setStatus(e.message || "Sign in failed");
    }
  }

  async function handleLogout() {
    try {
      await api.post("/auth/logout");
      setAccessToken(null);
      setIsAuthenticated(false);
      setProfile(null);
      setEmail("");
      setPassword("");
      setStatus("Logged out successfully ✅");
    } catch (err) {
      console.error(err);
      setStatus("Logout failed ❌");
    }
  }

  // ----------------------------
  // PROFILE PAGE
  // ----------------------------
  if (isAuthenticated && profile) {
    return (
      <div className="min-h-screen bg-gray-100 flex items-center justify-center">
        <div className="bg-white shadow-xl rounded-xl p-8 w-full max-w-md space-y-4">
          <h2 className="text-2xl font-semibold text-center">
            Profile Page 🔐
          </h2>

          <div className="bg-gray-50 p-4 rounded-lg">
            <p className="text-sm text-gray-500">User ID</p>
            <p className="font-medium break-all">{profile.user_id}</p>
          </div>

          <div className="bg-gray-50 p-4 rounded-lg">
            <p className="text-sm text-gray-500">Message</p>
            <p className="font-medium">{profile.message}</p>
          </div>

          <button
            onClick={handleLogout}
            className="w-full bg-red-600 hover:bg-red-700 text-white rounded-lg py-2 transition"
          >
            Logout
          </button>
        </div>
      </div>
    );
  }

  // ----------------------------
  // LOGIN PAGE
  // ----------------------------
  return (
    <div className="min-h-screen bg-gray-100 flex items-center justify-center">
      <div className="bg-white shadow-xl rounded-xl p-8 w-full max-w-md">
        <h2 className="text-2xl font-semibold text-center mb-6">
          Secure Auth Demo
        </h2>

        <div className="space-y-4">
          <input
            className="w-full border rounded-lg px-4 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />

          <input
            className="w-full border rounded-lg px-4 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="Password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />

          <div className="flex gap-3">
            <button
              onClick={handleSignUp}
              className="flex-1 bg-gray-200 hover:bg-gray-300 rounded-lg py-2 transition"
            >
              Sign Up
            </button>

            <button
              onClick={handleSignIn}
              className="flex-1 bg-blue-600 hover:bg-blue-700 text-white rounded-lg py-2 transition"
            >
              Sign In
            </button>
          </div>

          {status && (
            <p className="text-sm text-center text-gray-600 mt-3">
              {status}
            </p>
          )}
        </div>
      </div>
    </div>
  );
}
