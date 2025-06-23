// src/pages/CartPage.js
import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useDispatch, useSelector } from "react-redux";
import {
  Box,
  Container,
  Typography,
  Paper,
  Button,
  TextField,
  Divider,
  IconButton,
  List,
  ListItem,
  ListItemText,
  ListItemSecondaryAction,
  Alert,
} from "@mui/material";
import { Add, Remove, Delete, ShoppingCart } from "@mui/icons-material";
import { motion, AnimatePresence } from "framer-motion";
import {
  updateQuantity,
  removeFromCart,
  clearCart,
} from "../store/slices/cartSlice";
import { createOrder } from "../store/slices/orderSlice";
import { formatCurrency, calculateOrderTotals } from "../utils/helpers";
import EmptyState from "../components/EmptyState";

const CartPage = () => {
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const { items, itemCount } = useSelector((state) => state.cart);
  const { isLoading } = useSelector((state) => state.order);
  const [notes, setNotes] = useState("");

  const { subtotal, tax, serviceCharge, total } = calculateOrderTotals(items);

  const handleQuantityChange = (cartId, newQuantity) => {
    dispatch(updateQuantity({ cartId, quantity: newQuantity }));
  };

  const handleRemoveItem = (cartId) => {
    dispatch(removeFromCart({ cartId }));
  };

  const handleCheckout = async () => {
    const result = await dispatch(createOrder({ items, notes }));

    if (createOrder.fulfilled.match(result)) {
      dispatch(clearCart());
      navigate(`/orders/${result.payload.id}`);
    }
  };

  if (items.length === 0) {
    return (
      <Container maxWidth="sm" sx={{ py: 4 }}>
        <EmptyState
          icon={<ShoppingCart sx={{ fontSize: 80 }} />}
          title="Your cart is empty"
          subtitle="Add some delicious items from our menu"
          action={
            <Button variant="contained" onClick={() => navigate("/menu")}>
              Browse Menu
            </Button>
          }
        />
      </Container>
    );
  }

  return (
    <Container maxWidth="md" sx={{ py: 2 }}>
      <Typography variant="h4" gutterBottom fontWeight="bold">
        Your Cart ({itemCount} items)
      </Typography>

      <Grid container spacing={3}>
        <Grid item xs={12} md={8}>
          <Paper sx={{ p: 2 }}>
            <List>
              <AnimatePresence>
                {items.map((item) => (
                  <motion.div
                    key={item.cartId}
                    initial={{ opacity: 0, x: -20 }}
                    animate={{ opacity: 1, x: 0 }}
                    exit={{ opacity: 0, x: 20 }}
                    transition={{ duration: 0.3 }}>
                    <ListItem divider>
                      <ListItemText
                        primary={item.name}
                        secondary={
                          <>
                            {formatCurrency(item.price)} each
                            {item.notes && (
                              <Typography
                                variant="caption"
                                display="block"
                                color="text.secondary">
                                Note: {item.notes}
                              </Typography>
                            )}
                          </>
                        }
                      />
                      <ListItemSecondaryAction>
                        <Box display="flex" alignItems="center" gap={1}>
                          <IconButton
                            size="small"
                            onClick={() =>
                              handleQuantityChange(
                                item.cartId,
                                item.quantity - 1
                              )
                            }
                            disabled={item.quantity <= 1}>
                            <Remove />
                          </IconButton>
                          <Typography
                            sx={{ minWidth: 30, textAlign: "center" }}>
                            {item.quantity}
                          </Typography>
                          <IconButton
                            size="small"
                            onClick={() =>
                              handleQuantityChange(
                                item.cartId,
                                item.quantity + 1
                              )
                            }>
                            <Add />
                          </IconButton>
                          <Typography
                            sx={{
                              minWidth: 80,
                              textAlign: "right",
                              fontWeight: "bold",
                            }}>
                            {formatCurrency(item.price * item.quantity)}
                          </Typography>
                          <IconButton
                            size="small"
                            color="error"
                            onClick={() => handleRemoveItem(item.cartId)}>
                            <Delete />
                          </IconButton>
                        </Box>
                      </ListItemSecondaryAction>
                    </ListItem>
                  </motion.div>
                ))}
              </AnimatePresence>
            </List>

            <Box mt={3}>
              <TextField
                fullWidth
                multiline
                rows={3}
                label="Special Instructions"
                placeholder="Any special requests or dietary requirements?"
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
              />
            </Box>
          </Paper>
        </Grid>

        <Grid item xs={12} md={4}>
          <Paper sx={{ p: 3, position: "sticky", top: 80 }}>
            <Typography variant="h6" gutterBottom>
              Order Summary
            </Typography>

            <Box sx={{ my: 2 }}>
              <Box display="flex" justifyContent="space-between" mb={1}>
                <Typography>Subtotal</Typography>
                <Typography>{formatCurrency(subtotal)}</Typography>
              </Box>
              <Box display="flex" justifyContent="space-between" mb={1}>
                <Typography>Tax (10%)</Typography>
                <Typography>{formatCurrency(tax)}</Typography>
              </Box>
              <Box display="flex" justifyContent="space-between" mb={1}>
                <Typography>Service Charge (5%)</Typography>
                <Typography>{formatCurrency(serviceCharge)}</Typography>
              </Box>

              <Divider sx={{ my: 2 }} />

              <Box display="flex" justifyContent="space-between">
                <Typography variant="h6">Total</Typography>
                <Typography variant="h6" color="primary">
                  {formatCurrency(total)}
                </Typography>
              </Box>
            </Box>

            <Button
              fullWidth
              variant="contained"
              size="large"
              onClick={handleCheckout}
              disabled={isLoading}>
              Place Order
            </Button>

            <Alert severity="info" sx={{ mt: 2 }}>
              Payment will be processed after your order is confirmed
            </Alert>
          </Paper>
        </Grid>
      </Grid>
    </Container>
  );
};

export default CartPage;
