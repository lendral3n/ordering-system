// src/services/api.js
import axios from "axios";
import { store } from "../store/store";
import { clearSession } from "../store/slices/sessionSlice";
import { showSnackbar } from "../store/slices/notificationSlice";

const API_URL = process.env.REACT_APP_API_URL || "http://localhost:8080/api";

const api = axios.create({
  baseURL: API_URL,
  headers: {
    "Content-Type": "application/json",
  },
});

// Request interceptor
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem("session_token");
    if (token) {
      config.headers["X-Session-Token"] = token;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response) {
      // Handle specific error codes
      switch (error.response.status) {
        case 401:
          // Session expired or invalid
          store.dispatch(clearSession());
          store.dispatch(
            showSnackbar({
              message: "Session expired. Please scan the QR code again.",
              severity: "error",
            })
          );
          window.location.href = "/scan";
          break;
        case 500:
          store.dispatch(
            showSnackbar({
              message: "Server error. Please try again later.",
              severity: "error",
            })
          );
          break;
        default:
          const message = error.response.data?.error || "An error occurred";
          store.dispatch(
            showSnackbar({
              message,
              severity: "error",
            })
          );
      }
    } else if (error.request) {
      store.dispatch(
        showSnackbar({
          message: "Network error. Please check your connection.",
          severity: "error",
        })
      );
    }
    return Promise.reject(error);
  }
);

export default api;