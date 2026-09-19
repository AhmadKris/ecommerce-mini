import { Outlet, Route, Routes } from "react-router-dom";

import { AdminRoute } from "./components/AdminRoute";
import { ProtectedRoute } from "./components/ProtectedRoute";
import { AdminLayout } from "./components/layout/AdminLayout";
import { Header } from "./components/layout/Header";
import { Home } from "./pages/Home";
import { AdminCategories } from "./pages/admin/AdminCategories";
import { AdminCustomerDetail } from "./pages/admin/AdminCustomerDetail";
import { AdminCustomers } from "./pages/admin/AdminCustomers";
import { AdminDashboard } from "./pages/admin/AdminDashboard";
import { AdminInventory } from "./pages/admin/AdminInventory";
import { AdminOrderDetail } from "./pages/admin/AdminOrderDetail";
import { AdminOrders } from "./pages/admin/AdminOrders";
import { AdminProducts } from "./pages/admin/AdminProducts";
import { AdminPromotions } from "./pages/admin/AdminPromotions";
import { AdminReviews } from "./pages/admin/AdminReviews";
import { ProductForm } from "./pages/admin/ProductForm";
import { ForgotPassword } from "./pages/auth/ForgotPassword";
import { Login } from "./pages/auth/Login";
import { Register } from "./pages/auth/Register";
import { ResetPassword } from "./pages/auth/ResetPassword";
import { Cart } from "./pages/cart/Cart";
import { Checkout } from "./pages/checkout/Checkout";
import { OrderDetail } from "./pages/orders/OrderDetail";
import { OrderHistory } from "./pages/orders/OrderHistory";
import { Profile } from "./pages/profile/Profile";
import { ProductDetail } from "./pages/products/ProductDetail";
import { ProductList } from "./pages/products/ProductList";

function Layout() {
  return (
    <>
      <Header />
      <Outlet />
    </>
  );
}

function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<Home />} />
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/forgot-password" element={<ForgotPassword />} />
        <Route path="/reset-password" element={<ResetPassword />} />
        <Route path="/products" element={<ProductList />} />
        <Route path="/products/:slug" element={<ProductDetail />} />
        <Route element={<ProtectedRoute />}>
          <Route path="/cart" element={<Cart />} />
          <Route path="/checkout" element={<Checkout />} />
          <Route path="/orders" element={<OrderHistory />} />
          <Route path="/orders/:id" element={<OrderDetail />} />
          <Route path="/profile" element={<Profile />} />
        </Route>
      </Route>
      <Route element={<AdminRoute />}>
        <Route element={<AdminLayout />}>
          <Route path="/admin" element={<AdminDashboard />} />
          <Route path="/admin/products" element={<AdminProducts />} />
          <Route path="/admin/products/new" element={<ProductForm />} />
          <Route path="/admin/products/:id/edit" element={<ProductForm />} />
          <Route path="/admin/categories" element={<AdminCategories />} />
          <Route path="/admin/orders" element={<AdminOrders />} />
          <Route path="/admin/orders/:id" element={<AdminOrderDetail />} />
          <Route path="/admin/customers" element={<AdminCustomers />} />
          <Route path="/admin/customers/:id" element={<AdminCustomerDetail />} />
          <Route path="/admin/inventory" element={<AdminInventory />} />
          <Route path="/admin/promotions" element={<AdminPromotions />} />
          <Route path="/admin/reviews" element={<AdminReviews />} />
        </Route>
      </Route>
    </Routes>
  );
}

export default App;
