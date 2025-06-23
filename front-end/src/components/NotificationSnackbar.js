// src/components/NotificationSnackbar.js
import React from "react";
import { useSelector, useDispatch } from "react-redux";
import { Snackbar, Alert } from "@mui/material";
import { hideSnackbar } from "../store/slices/notificationSlice";

const NotificationSnackbar = () => {
  const dispatch = useDispatch();
  const { snackbar } = useSelector((state) => state.notification);

  const handleClose = (event, reason) => {
    if (reason === "clickaway") {
      return;
    }
    dispatch(hideSnackbar());
  };

  return (
    <Snackbar
      open={snackbar.open}
      autoHideDuration={4000}
      onClose={handleClose}
      anchorOrigin={{ vertical: "bottom", horizontal: "center" }}>
      <Alert
        onClose={handleClose}
        severity={snackbar.severity}
        sx={{ width: "100%" }}
        elevation={6}
        variant="filled">
        {snackbar.message}
      </Alert>
    </Snackbar>
  );
};

export default NotificationSnackbar;
