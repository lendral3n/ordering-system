// src/components/Menu/MenuItemDialog.js
import React, { useState } from "react";
import { useDispatch } from "react-redux";
import {
  Dialog,
  DialogContent,
  DialogActions,
  Typography,
  Box,
  Button,
  TextField,
  IconButton,
  Chip,
  Tab,
  Tabs,
} from "@mui/material";
import { Close, Add, Remove, Image360, PlayCircle } from "@mui/icons-material";
import { addToCart } from "../../store/slices/cartSlice";
import { showSnackbar } from "../../store/slices/notificationSlice";
import { formatCurrency } from "../../utils/helpers";
import Product360View from "./Product360View";

const MenuItemDialog = ({ item, open, onClose }) => {
  const dispatch = useDispatch();
  const [quantity, setQuantity] = useState(1);
  const [notes, setNotes] = useState("");
  const [activeTab, setActiveTab] = useState(0);

  if (!item) return null;

  const handleAddToCart = () => {
    dispatch(addToCart({ item, quantity, notes }));
    dispatch(
      showSnackbar({
        message: `${item.name} added to cart`,
        severity: "success",
      })
    );
    onClose();
    setQuantity(1);
    setNotes("");
  };

  const handleQuantityChange = (delta) => {
    const newQuantity = quantity + delta;
    if (newQuantity >= 1) {
      setQuantity(newQuantity);
    }
  };

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="md"
      fullWidth
      PaperProps={{
        sx: { borderRadius: 2 },
      }}>
      <Box sx={{ position: "relative" }}>
        <IconButton
          onClick={onClose}
          sx={{
            position: "absolute",
            right: 8,
            top: 8,
            zIndex: 1,
          }}>
          <Close />
        </IconButton>

        {/* Media Tabs */}
        {(item.image_360_url || item.video_url) && (
          <Tabs
            value={activeTab}
            onChange={(e, newValue) => setActiveTab(newValue)}
            sx={{ borderBottom: 1, borderColor: "divider" }}>
            <Tab label="Image" />
            {item.image_360_url && (
              <Tab label="360° View" icon={<Image360 />} />
            )}
            {item.video_url && <Tab label="Video" icon={<PlayCircle />} />}
          </Tabs>
        )}

        {/* Media Content */}
        <Box
          sx={{ height: 300, overflow: "hidden", backgroundColor: "#f5f5f5" }}>
          {activeTab === 0 && (
            <img
              src={item.image_url || "/placeholder-food.jpg"}
              alt={item.name}
              style={{
                width: "100%",
                height: "100%",
                objectFit: "cover",
              }}
            />
          )}
          {activeTab === 1 && item.image_360_url && (
            <Product360View imageUrl={item.image_360_url} />
          )}
          {activeTab === 2 && item.video_url && (
            <video
              src={item.video_url}
              controls
              style={{
                width: "100%",
                height: "100%",
                objectFit: "cover",
              }}
            />
          )}
        </Box>
      </Box>

      <DialogContent>
        <Typography variant="h5" gutterBottom>
          {item.name}
        </Typography>

        <Typography variant="body1" color="text.secondary" paragraph>
          {item.description}
        </Typography>

        <Box display="flex" gap={1} mb={2}>
          {item.preparation_time && (
            <Chip
              label={`~${item.preparation_time} mins`}
              size="small"
              variant="outlined"
            />
          )}
          {item.stock_quantity && (
            <Chip
              label={`${item.stock_quantity} available`}
              size="small"
              variant="outlined"
              color={item.stock_quantity < 10 ? "warning" : "default"}
            />
          )}
        </Box>

        <TextField
          fullWidth
          multiline
          rows={2}
          label="Special Instructions"
          placeholder="Any special requests?"
          value={notes}
          onChange={(e) => setNotes(e.target.value)}
          sx={{ mb: 3 }}
        />

        <Box display="flex" justifyContent="space-between" alignItems="center">
          <Typography variant="h6">
            {formatCurrency(item.price * quantity)}
          </Typography>

          <Box display="flex" alignItems="center" gap={2}>
            <IconButton
              onClick={() => handleQuantityChange(-1)}
              disabled={quantity <= 1}>
              <Remove />
            </IconButton>
            <Typography variant="h6" sx={{ minWidth: 30, textAlign: "center" }}>
              {quantity}
            </Typography>
            <IconButton
              onClick={() => handleQuantityChange(1)}
              disabled={item.stock_quantity && quantity >= item.stock_quantity}>
              <Add />
            </IconButton>
          </Box>
        </Box>
      </DialogContent>

      <DialogActions sx={{ px: 3, pb: 3 }}>
        <Button onClick={onClose} size="large">
          Cancel
        </Button>
        <Button
          variant="contained"
          onClick={handleAddToCart}
          size="large"
          disabled={!item.is_available}
          startIcon={<Add />}>
          Add to Cart
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default MenuItemDialog;
