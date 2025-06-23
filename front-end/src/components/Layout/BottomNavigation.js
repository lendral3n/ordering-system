// src/components/Layout/BottomNavigation.js
import React from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { useSelector } from "react-redux";
import {
  BottomNavigation as MuiBottomNavigation,
  BottomNavigationAction,
  Badge,
  Paper,
} from "@mui/material";
import {
  RestaurantMenu,
  ShoppingCart,
  Receipt,
  History,
} from "@mui/icons-material";

const BottomNavigation = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { itemCount } = useSelector((state) => state.cart);

  const routes = [
    { label: "Menu", value: "/menu", icon: <RestaurantMenu /> },
    {
      label: "Cart",
      value: "/cart",
      icon: (
        <Badge badgeContent={itemCount} color="error">
          <ShoppingCart />
        </Badge>
      ),
    },
    { label: "Orders", value: "/orders", icon: <Receipt /> },
    { label: "History", value: "/orders/history", icon: <History /> },
  ];

  const currentRoute =
    routes.find((route) => location.pathname.startsWith(route.value))?.value ||
    "/menu";

  return (
    <Paper
      sx={{
        position: "fixed",
        bottom: 0,
        left: 0,
        right: 0,
        zIndex: 1000,
      }}
      elevation={3}>
      <MuiBottomNavigation
        value={currentRoute}
        onChange={(event, newValue) => {
          navigate(newValue);
        }}>
        {routes.map((route) => (
          <BottomNavigationAction
            key={route.value}
            label={route.label}
            value={route.value}
            icon={route.icon}
          />
        ))}
      </MuiBottomNavigation>
    </Paper>
  );
};

export default BottomNavigation;
