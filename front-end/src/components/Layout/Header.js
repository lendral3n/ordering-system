// src/components/Layout/Header.js
import React, { useState } from "react";
import { useSelector, useDispatch } from "react-redux";
import { useNavigate } from "react-router-dom";
import {
  AppBar,
  Toolbar,
  Typography,
  IconButton,
  Badge,
  Menu,
  MenuItem,
  Box,
  Chip,
} from "@mui/material";
import { Notifications, ExitToApp, TableRestaurant } from "@mui/icons-material";
import { endSession } from "../../store/slices/sessionSlice";

const Header = () => {
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const { table } = useSelector((state) => state.session);
  const { notifications } = useSelector((state) => state.notification);
  const [anchorEl, setAnchorEl] = useState(null);

  const unreadCount = notifications.filter((n) => !n.read).length;

  const handleNotificationClick = (event) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const handleLogout = async () => {
    await dispatch(endSession());
    navigate("/scan");
  };

  return (
    <AppBar position="sticky" elevation={1}>
      <Toolbar>
        <Box
          sx={{ flexGrow: 1, display: "flex", alignItems: "center", gap: 2 }}>
          <Typography variant="h6" component="div">
            Restaurant
          </Typography>
          {table && (
            <Chip
              icon={<TableRestaurant />}
              label={`Table ${table.table_number}`}
              color="secondary"
              size="small"
            />
          )}
        </Box>

        <IconButton color="inherit" onClick={handleNotificationClick}>
          <Badge badgeContent={unreadCount} color="error">
            <Notifications />
          </Badge>
        </IconButton>

        <IconButton color="inherit" onClick={handleLogout}>
          <ExitToApp />
        </IconButton>
      </Toolbar>

      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={handleClose}
        PaperProps={{
          style: {
            maxHeight: 400,
            width: "300px",
          },
        }}>
        {notifications.length === 0 ? (
          <MenuItem disabled>No notifications</MenuItem>
        ) : (
          notifications.map((notification) => (
            <MenuItem
              key={notification.id}
              onClick={handleClose}
              sx={{
                backgroundColor: notification.read
                  ? "transparent"
                  : "action.hover",
              }}>
              <Box>
                <Typography variant="subtitle2">
                  {notification.title}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  {notification.message}
                </Typography>
              </Box>
            </MenuItem>
          ))
        )}
      </Menu>
    </AppBar>
  );
};

export default Header;
