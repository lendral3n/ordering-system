// src/pages/OrderTrackingPage.js
import React, { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useDispatch, useSelector } from "react-redux";
import {
  Box,
  Container,
  Typography,
  Paper,
  Stepper,
  Step,
  StepLabel,
  StepContent,
  Button,
  Chip,
  Grid,
  List,
  ListItem,
  ListItemText,
  Divider,
  LinearProgress,
  Alert,
} from "@mui/material";
import {
  Timer,
  CheckCircle,
  Payment,
  HelpOutline,
} from "@mui/icons-material";
import { motion } from "framer-motion";
import { fetchOrder, requestAssistance } from "../store/slices/orderSlice";
import { formatCurrency, formatDate, getOrderProgress } from "../utils/helpers";
import {
  ORDER_STATUS,
  ORDER_STATUS_LABELS,
  ORDER_STATUS_COLORS,
  PAYMENT_STATUS_LABELS,
} from "../utils/constants";

const OrderTrackingPage = () => {
  const { orderId } = useParams();
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const { currentOrder, isLoading } = useSelector((state) => state.order);
  const [assistanceRequested, setAssistanceRequested] = useState(false);

  useEffect(() => {
    const fetchOrderData = () => {
      dispatch(fetchOrder(orderId));
    };

    fetchOrderData();

    // Refresh order status every 10 seconds
    const interval = setInterval(fetchOrderData, 10000);

    return () => clearInterval(interval);
  }, [dispatch, orderId]);

  const handleRequestAssistance = async () => {
    await dispatch(requestAssistance());
    setAssistanceRequested(true);
  };

  const handlePayment = () => {
    navigate(`/payment/${orderId}`);
  };

  const getActiveStep = () => {
    if (!currentOrder) return 0;

    const steps = [
      ORDER_STATUS.PENDING,
      ORDER_STATUS.CONFIRMED,
      ORDER_STATUS.PREPARING,
      ORDER_STATUS.READY,
      ORDER_STATUS.SERVED,
    ];

    const currentIndex = steps.indexOf(currentOrder.status);
    return currentIndex >= 0 ? currentIndex : 0;
  };

  if (isLoading && !currentOrder) {
    return (
      <Container maxWidth="md" sx={{ py: 4 }}>
        <Paper sx={{ p: 3 }}>
          <LinearProgress />
        </Paper>
      </Container>
    );
  }

  if (!currentOrder) {
    return (
      <Container maxWidth="md" sx={{ py: 4 }}>
        <Alert severity="error">Order not found</Alert>
      </Container>
    );
  }

  const progress = getOrderProgress(currentOrder.status);

  return (
    <Container maxWidth="md" sx={{ py: 3 }}>
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}>
        <Paper sx={{ p: 3, mb: 3 }}>
          <Box
            display="flex"
            justifyContent="space-between"
            alignItems="center"
            mb={3}>
            <Box>
              <Typography variant="h4" gutterBottom>
                Order #{currentOrder.order_number}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                {formatDate(currentOrder.created_at)}
              </Typography>
            </Box>
            <Box textAlign="right">
              <Chip
                label={ORDER_STATUS_LABELS[currentOrder.status]}
                color={ORDER_STATUS_COLORS[currentOrder.status]}
                sx={{ mb: 1 }}
              />
              <Typography variant="body2" color="text.secondary">
                {PAYMENT_STATUS_LABELS[currentOrder.payment_status]}
              </Typography>
            </Box>
          </Box>

          <LinearProgress
            variant="determinate"
            value={progress}
            sx={{ height: 8, borderRadius: 4, mb: 3 }}
          />

          <Stepper activeStep={getActiveStep()} orientation="vertical">
            <Step>
              <StepLabel>
                <Typography>Order Placed</Typography>
              </StepLabel>
              <StepContent>
                <Typography variant="body2" color="text.secondary">
                  Your order has been received
                </Typography>
              </StepContent>
            </Step>

            <Step>
              <StepLabel>
                <Typography>Order Confirmed</Typography>
              </StepLabel>
              <StepContent>
                <Typography variant="body2" color="text.secondary">
                  Kitchen has confirmed your order
                </Typography>
              </StepContent>
            </Step>

            <Step>
              <StepLabel>
                <Typography>Preparing</Typography>
              </StepLabel>
              <StepContent>
                <Typography variant="body2" color="text.secondary">
                  Your food is being prepared
                </Typography>
              </StepContent>
            </Step>

            <Step>
              <StepLabel>
                <Typography>Ready</Typography>
              </StepLabel>
              <StepContent>
                <Typography variant="body2" color="text.secondary">
                  Your order is ready to be served
                </Typography>
              </StepContent>
            </Step>

            <Step>
              <StepLabel>
                <Typography>Served</Typography>
              </StepLabel>
              <StepContent>
                <Typography variant="body2" color="text.secondary">
                  Enjoy your meal!
                </Typography>
              </StepContent>
            </Step>
          </Stepper>
        </Paper>

        <Grid container spacing={3}>
          <Grid item xs={12} md={8}>
            <Paper sx={{ p: 3 }}>
              <Typography variant="h6" gutterBottom>
                Order Items
              </Typography>

              <List>
                {currentOrder.order_items?.map((item) => (
                  <ListItem key={item.id} divider>
                    <ListItemText
                      primary={item.menu_item?.name}
                      secondary={
                        <>
                          {item.quantity} x {formatCurrency(item.unit_price)}
                          {item.notes && (
                            <Typography variant="caption" display="block">
                              Note: {item.notes}
                            </Typography>
                          )}
                        </>
                      }
                    />
                    <Typography fontWeight="bold">
                      {formatCurrency(item.subtotal)}
                    </Typography>
                  </ListItem>
                ))}
              </List>

              <Divider sx={{ my: 2 }} />

              <Box>
                <Box display="flex" justifyContent="space-between" mb={1}>
                  <Typography>Subtotal</Typography>
                  <Typography>
                    {formatCurrency(currentOrder.total_amount)}
                  </Typography>
                </Box>
                <Box display="flex" justifyContent="space-between" mb={1}>
                  <Typography>Tax</Typography>
                  <Typography>
                    {formatCurrency(currentOrder.tax_amount)}
                  </Typography>
                </Box>
                <Box display="flex" justifyContent="space-between" mb={1}>
                  <Typography>Service Charge</Typography>
                  <Typography>
                    {formatCurrency(currentOrder.service_charge)}
                  </Typography>
                </Box>
                <Divider sx={{ my: 1 }} />
                <Box display="flex" justifyContent="space-between">
                  <Typography variant="h6">Total</Typography>
                  <Typography variant="h6" color="primary">
                    {formatCurrency(currentOrder.grand_total)}
                  </Typography>
                </Box>
              </Box>
            </Paper>
          </Grid>

          <Grid item xs={12} md={4}>
            <Paper sx={{ p: 3, mb: 2 }}>
              <Typography variant="h6" gutterBottom>
                Actions
              </Typography>

              {currentOrder.payment_status === "unpaid" && (
                <Button
                  fullWidth
                  variant="contained"
                  startIcon={<Payment />}
                  onClick={handlePayment}
                  sx={{ mb: 2 }}>
                  Pay Now
                </Button>
              )}

              <Button
                fullWidth
                variant="outlined"
                startIcon={<HelpOutline />}
                onClick={handleRequestAssistance}
                disabled={assistanceRequested}>
                {assistanceRequested ? "Assistance Requested" : "Call Waiter"}
              </Button>
            </Paper>

            {currentOrder.status === ORDER_STATUS.PREPARING && (
              <Alert severity="info" icon={<Timer />}>
                Estimated preparation time: 20-30 minutes
              </Alert>
            )}

            {currentOrder.status === ORDER_STATUS.READY && (
              <Alert severity="success" icon={<CheckCircle />}>
                Your order is ready! Our staff will serve it shortly.
              </Alert>
            )}
          </Grid>
        </Grid>
      </motion.div>
    </Container>
  );
};

export default OrderTrackingPage;
