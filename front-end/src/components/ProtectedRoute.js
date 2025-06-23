// src/components/ProtectedRoute.js
import React from "react";
import { Navigate, Outlet } from "react-router-dom";
import { useSelector } from "react-redux";

const ProtectedRoute = () => {
  const { isAuthenticated } = useSelector((state) => state.session);

  if (!isAuthenticated) {
    return <Navigate to="/scan" replace />;
  }

  return <Outlet />;
};

export default ProtectedRoute;
