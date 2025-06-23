// src/pages/OrderHistoryPage.js
import React, { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useDispatch, useSelector } from "react-redux";
import {
  Container,
  Typography,
  Paper,
  List,
  ListItem,
  ListItemText,
  ListItemSecondaryAction,
  Chip,
  Button,
  Box,
} from "@mui/material";
import { Receipt, ChevronRight } from "@mui/icons-material";
import { fetchOrderHistory } from "../store/slices/orderSlice";
import { formatCurrency, formatDate } from "../utils/helpers";
import { ORDER_STATUS_LABELS, ORDER_STATUS_COLORS } from "../utils/constants";
import EmptyState from "../components/EmptyState";

const OrderHistoryPage = () => {
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const { orderHistory } = useSelector((state) => state.order);

  useEffect(() => {
    dispatch(fetchOrderHistory());
  }, [dispatch]);

  if (orderHistory.length === 0) {
    return (
      <Container maxWidth="sm" sx={{ py: 4 }}>
        <EmptyState
          icon={<Receipt sx={{ fontSize: 80 }} />}
          title="No orders yet"
          subtitle="Your order history will appear here"
          action={
            <Button variant="contained" onClick={() => navigate("/menu")}>
              Start Ordering
            </Button>
          }
        />
      </Container>
    );
  }

  return (
    <Container maxWidth="md" sx={{ py: 3 }}>
      <Typography variant="h4" gutterBottom>
        Order History
      </Typography>

      <List>
        {orderHistory.map((order) => (
          <Paper key={order.id} sx={{ mb: 2 }}>
            <ListItem button onClick={() => navigate(`/orders/${order.id}`)}>
              <ListItemText
                primary={
                  <Box display="flex" alignItems="center" gap={2}>
                    <Typography variant="h6">
                      Order #{order.order_number}
                    </Typography>
                    <Chip
                      label={ORDER_STATUS_LABELS[order.status]}
                      color={ORDER_STATUS_COLORS[order.status]}
                      size="small"
                    />
                  </Box>
                }
                secondary={
                  <>
                    <Typography variant="body2" color="text.secondary">
                      {formatDate(order.created_at)}
                    </Typography>
                    <Typography variant="body2">
                      {order.order_items?.length || 0} items •{" "}
                      {formatCurrency(order.grand_total)}
                    </Typography>
                  </>
                }
              />
              <ListItemSecondaryAction>
                <ChevronRight />
              </ListItemSecondaryAction>
            </ListItem>
          </Paper>
        ))}
      </List>
    </Container>
  );
};

export default OrderHistoryPage;
