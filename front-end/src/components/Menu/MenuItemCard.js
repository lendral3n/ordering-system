// src/components/Menu/MenuItemCard.js
import React from "react";
import {
  Card,
  CardMedia,
  CardContent,
  Typography,
  Box,
  Chip,
} from "@mui/material";
import { Image360, PlayCircle } from "@mui/icons-material";
import { formatCurrency } from "../../utils/helpers";
import { motion } from "framer-motion";

const MenuItemCard = ({ item, onClick }) => {
  const handleClick = () => {
    onClick(item);
  };

  return (
    <motion.div whileHover={{ scale: 1.02 }} whileTap={{ scale: 0.98 }}>
      <Card
        sx={{
          height: "100%",
          cursor: "pointer",
          position: "relative",
          "&:hover": {
            boxShadow: 4,
          },
        }}
        onClick={handleClick}>
        <CardMedia
          component="img"
          height="200"
          image={item.image_url || "/placeholder-food.jpg"}
          alt={item.name}
          sx={{ objectFit: "cover" }}
        />

        {/* Media badges */}
        <Box
          sx={{
            position: "absolute",
            top: 8,
            right: 8,
            display: "flex",
            gap: 1,
          }}>
          {item.image_360_url && (
            <Chip
              icon={<Image360 />}
              label="360°"
              size="small"
              sx={{ backgroundColor: "rgba(255, 255, 255, 0.9)" }}
            />
          )}
          {item.video_url && (
            <Chip
              icon={<PlayCircle />}
              label="Video"
              size="small"
              sx={{ backgroundColor: "rgba(255, 255, 255, 0.9)" }}
            />
          )}
        </Box>

        <CardContent>
          <Typography variant="h6" gutterBottom noWrap>
            {item.name}
          </Typography>

          <Typography
            variant="body2"
            color="text.secondary"
            sx={{
              mb: 2,
              overflow: "hidden",
              textOverflow: "ellipsis",
              display: "-webkit-box",
              WebkitLineClamp: 2,
              WebkitBoxOrient: "vertical",
            }}>
            {item.description}
          </Typography>

          <Box
            display="flex"
            justifyContent="space-between"
            alignItems="center">
            <Typography variant="h6" color="primary">
              {formatCurrency(item.price)}
            </Typography>

            {!item.is_available && (
              <Chip label="Unavailable" size="small" color="error" />
            )}
          </Box>

          {item.preparation_time && (
            <Typography variant="caption" color="text.secondary">
              ~{item.preparation_time} mins
            </Typography>
          )}
        </CardContent>
      </Card>
    </motion.div>
  );
};

export default MenuItemCard;
