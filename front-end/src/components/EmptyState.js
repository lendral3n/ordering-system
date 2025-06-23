// src/components/EmptyState.js
import React from "react";
import { Box, Typography } from "@mui/material";
import { motion } from "framer-motion";

const EmptyState = ({ icon, title, subtitle, action }) => {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5 }}>
      <Box
        sx={{
          textAlign: "center",
          py: 8,
          px: 2,
        }}>
        <Box sx={{ color: "text.secondary", mb: 2 }}>{icon}</Box>
        <Typography variant="h5" gutterBottom>
          {title}
        </Typography>
        <Typography variant="body1" color="text.secondary" sx={{ mb: 4 }}>
          {subtitle}
        </Typography>
        {action}
      </Box>
    </motion.div>
  );
};

export default EmptyState;
