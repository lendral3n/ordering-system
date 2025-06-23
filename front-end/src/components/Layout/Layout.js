// src/components/Layout/Layout.js
import React from 'react';
import { Outlet } from 'react-router-dom';
import { Box } from '@mui/material';
import Header from './Header';
import BottomNavigation from './BottomNavigation';
import NotificationSnackbar from '../NotificationSnackbar';

const Layout = () => {
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
      <Header />
      <Box component="main" sx={{ flexGrow: 1, pb: 8 }}>
        <Outlet />
      </Box>
      <BottomNavigation />
      <NotificationSnackbar />
    </Box>
  );
};

export default Layout;