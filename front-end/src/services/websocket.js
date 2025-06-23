// src/services/websocket.js
import io from "socket.io-client";
import { store } from "../store/store";
import {
  addNotification,
  showSnackbar,
} from "../store/slices/notificationSlice";
import { updateOrderStatus } from "../store/slices/orderSlice";

const WS_URL = process.env.REACT_APP_WS_URL || "ws://localhost:8080";

let socket = null;

export const connectWebSocket = () => {
  const state = store.getState();
  const { table } = state.session;

  if (!table) return;

  socket = io(WS_URL, {
    query: {
      role: "customer",
      table_id: table.id,
    },
  });

  socket.on("connect", () => {
    console.log("WebSocket connected");
  });

  socket.on("disconnect", () => {
    console.log("WebSocket disconnected");
  });

  // Handle notifications
  socket.on("order_status_updated", (data) => {
    store.dispatch(
      updateOrderStatus({
        orderId: data.order_id,
        status: data.status,
      })
    );

    store.dispatch(
      addNotification({
        type: "order_status",
        title: "Order Update",
        message: data.message,
        data,
      })
    );

    store.dispatch(
      showSnackbar({
        message: data.message,
        severity: "info",
      })
    );
  });

  socket.on("order_ready", (data) => {
    store.dispatch(
      addNotification({
        type: "order_ready",
        title: "Order Ready!",
        message: "Your order is ready to be served!",
        data,
      })
    );

    store.dispatch(
      showSnackbar({
        message: "Your order is ready! 🎉",
        severity: "success",
      })
    );

    // Play notification sound
    playNotificationSound();
  });

  socket.on("payment_confirmed", (data) => {
    store.dispatch(
      addNotification({
        type: "payment_confirmed",
        title: "Payment Confirmed",
        message: "Your payment has been confirmed!",
        data,
      })
    );

    store.dispatch(
      showSnackbar({
        message: "Payment confirmed! Thank you.",
        severity: "success",
      })
    );
  });
};

export const disconnectWebSocket = () => {
  if (socket) {
    socket.disconnect();
    socket = null;
  }
};

const playNotificationSound = () => {
  const audio = new Audio("/notification.mp3");
  audio.play().catch((error) => {
    console.error("Failed to play notification sound:", error);
  });
};
