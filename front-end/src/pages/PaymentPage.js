// src/pages/PaymentPage.js
import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import {
  Box,
  Container,
  Typography,
  Paper,
  Button,
  TextField,
  CircularProgress,
  Alert,
  Stepper,
  Step,
  StepLabel,
  List,
  ListItem,
  ListItemText,
  Divider,
} from '@mui/material';
import { Payment, CheckCircle } from '@mui/icons-material';
import { createPayment, checkPaymentStatus } from '../store/slices/orderSlice';
import { formatCurrency } from '../utils/helpers';
import { payWithMidtrans } from '../services/payment';

const PaymentPage = () => {
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const { orderId } = useParams();
  const { currentOrder, paymentData, isLoading, error } = useSelector((state) => state.order);
  const [email, setEmail] = useState('');
  const [paymentStep, setPaymentStep] = useState(0);

  useEffect(() => {
    if (!currentOrder || currentOrder.id !== parseInt(orderId)) {
      navigate('/orders');
    }
  }, [currentOrder, orderId, navigate]);

  useEffect(() => {
    // Check payment status periodically
    const interval = setInterval(() => {
      if (currentOrder && currentOrder.payment_status === 'pending') {
        dispatch(checkPaymentStatus(currentOrder.id));
      }
    }, 5000);

    return () => clearInterval(interval);
  }, [dispatch, currentOrder]);

  const handleCreatePayment = async () => {
    setPaymentStep(1);
    
    const result = await dispatch(createPayment({
      orderId: currentOrder.id,
      customerEmail: email,
    }));

    if (createPayment.fulfilled.match(result)) {
      setPaymentStep(2);
      
      try {
        await payWithMidtrans(result.payload.token);
        setPaymentStep(3);
        
        // Check payment status
        setTimeout(() => {
          dispatch(checkPaymentStatus(currentOrder.id));
        }, 2000);
      } catch (error) {
        console.error('Payment failed:', error);
        setPaymentStep(1);
      }
    }
  };

  const steps = ['Order Details', 'Create Payment', 'Process Payment', 'Completed'];

  if (!currentOrder) {
    return null;
  }

  if (currentOrder.payment_status === 'paid') {
    return (
      <Container maxWidth="sm" sx={{ py: 4, textAlign: 'center' }}>
        <CheckCircle sx={{ fontSize: 80, color: 'success.main', mb: 2 }} />
        <Typography variant="h4" gutterBottom>
          Payment Successful!
        </Typography>
        <Typography variant="body1" color="text.secondary" gutterBottom>
          Thank you for your payment. Your order is being prepared.
        </Typography>
        <Button
          variant="contained"
          onClick={() => navigate(`/orders/${currentOrder.id}`)}
          sx={{ mt: 3 }}
        >
          Track Your Order
        </Button>
      </Container>
    );
  }

  return (
    <Container maxWidth="md" sx={{ py: 3 }}>
      <Typography variant="h4" gutterBottom>
        Payment
      </Typography>

      <Stepper activeStep={paymentStep} sx={{ mb: 4 }}>
        {steps.map((label) => (
          <Step key={label}>
            <StepLabel>{label}</StepLabel>
          </Step>
        ))}
      </Stepper>

      <Grid container spacing={3}>
        <Grid item xs={12} md={8}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Order Details
            </Typography>
            
            <Box display="flex" justifyContent="space-between" mb={2}>
              <Typography variant="body2" color="text.secondary">
                Order Number
              </Typography>
              <Typography fontWeight="bold">
                {currentOrder.order_number}
              </Typography>
            </Box>

            <Divider sx={{ my: 2 }} />

            <List dense>
              {currentOrder.order_items?.map((item) => (
                <ListItem key={item.id}>
                  <ListItemText
                    primary={item.menu_item?.name}
                    secondary={`${item.quantity} x ${formatCurrency(item.unit_price)}`}
                  />
                  <Typography>
                    {formatCurrency(item.subtotal)}
                  </Typography>
                </ListItem>
              ))}
            </List>

            <Divider sx={{ my: 2 }} />

            <Box>
              <Box display="flex" justifyContent="space-between" mb={1}>
                <Typography variant="body2">Subtotal</Typography>
                <Typography>{formatCurrency(currentOrder.total_amount)}</Typography>
              </Box>
              <Box display="flex" justifyContent="space-between" mb={1}>
                <Typography variant="body2">Tax (10%)</Typography>
                <Typography>{formatCurrency(currentOrder.tax_amount)}</Typography>
              </Box>
              <Box display="flex" justifyContent="space-between" mb={1}>
                <Typography variant="body2">Service Charge (5%)</Typography>
                <Typography>{formatCurrency(currentOrder.service_charge)}</Typography>
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
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Payment Information
            </Typography>

            {error && (
              <Alert severity="error" sx={{ mb: 2 }}>
                {error}
              </Alert>
            )}

            {paymentStep === 0 && (
              <Box>
                <TextField
                  fullWidth
                  type="email"
                  label="Email (optional)"
                  placeholder="For payment receipt"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  sx={{ mb: 3 }}
                />
                
                <Button
                  fullWidth
                  variant="contained"
                  size="large"
                  startIcon={<Payment />}
                  onClick={handleCreatePayment}
                  disabled={isLoading}
                >
                  Proceed to Payment
                </Button>

                <Alert severity="info" sx={{ mt: 2 }}>
                  You will be redirected to secure payment gateway
                </Alert>
              </Box>
            )}

            {paymentStep >= 1 && paymentStep < 3 && (
              <Box textAlign="center" py={3}>
                <CircularProgress />
                <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>
                  {paymentStep === 1 ? 'Creating payment...' : 'Processing payment...'}
                </Typography>
              </Box>
            )}

            {paymentStep === 3 && (
              <Box textAlign="center" py={3}>
                <CircularProgress />
                <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>
                  Verifying payment...
                </Typography>
              </Box>
            )}
          </Paper>
        </Grid>
      </Grid>
    </Container>
  );
};

export default PaymentPage;

