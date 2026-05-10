import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { User, Lock, Mail, ArrowRight, Globe, Cpu } from 'lucide-react';

const AuthPage: React.FC = () => {
  const [isLogin, setIsLogin] = useState(true);

  const toggleMode = () => setIsLogin(!isLogin);

  const containerVariants: any = {
    hidden: { opacity: 0, y: 10 },
    visible: { 
      opacity: 1, 
      y: 0,
      transition: { duration: 0.4, ease: "easeOut" }
    },
    exit: { 
      opacity: 0, 
      scale: 0.98,
      transition: { duration: 0.2 }
    }
  };

  return (
    <div className="min-h-screen bg-white flex items-center justify-center p-4 selection:bg-klein-blue selection:text-white">
      <motion.div 
        initial="hidden"
        animate="visible"
        variants={containerVariants}
        className="w-full max-w-[440px] bg-white rounded-2xl p-10 shadow-[0_8px_30px_rgb(0,0,0,0.04)] border border-gray-100"
      >
        <div className="text-center mb-10">
          <div className="inline-flex items-center justify-center w-12 h-12 rounded-xl bg-klein-blue/5 mb-6">
            <div className="w-6 h-6 rounded-full border-4 border-klein-blue" />
          </div>
          <motion.h1 
            key={isLogin ? 'login-h1' : 'register-h1'}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="text-2xl font-semibold text-gray-900 tracking-tight"
          >
            {isLogin ? 'Sign in to Minor' : 'Create your account'}
          </motion.h1>
          <p className="text-gray-500 mt-2 text-sm">
            {isLogin ? "Welcome back! Please enter your details." : "Start your journey with us today."}
          </p>
        </div>

        <form className="space-y-5" onSubmit={(e) => e.preventDefault()}>
          <AnimatePresence mode="wait">
            {!isLogin && (
              <motion.div
                initial={{ opacity: 0, height: 0 }}
                animate={{ opacity: 1, height: 'auto' }}
                exit={{ opacity: 0, height: 0 }}
                className="space-y-1.5"
              >
                <label className="text-sm font-medium text-gray-700">Email address</label>
                <div className="relative">
                  <Mail className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4.5 h-4.5 text-gray-400" />
                  <input 
                    type="email" 
                    className="w-full bg-white border border-gray-200 rounded-xl p-3 pl-11 text-gray-900 focus:ring-2 focus:ring-klein-blue/20 focus:border-klein-blue transition-all outline-none placeholder:text-gray-400"
                    placeholder="name@example.com"
                  />
                </div>
              </motion.div>
            )}
          </AnimatePresence>

          <div className="space-y-1.5">
            <label className="text-sm font-medium text-gray-700">Username</label>
            <div className="relative">
              <User className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4.5 h-4.5 text-gray-400" />
              <input 
                type="text" 
                className="w-full bg-white border border-gray-200 rounded-xl p-3 pl-11 text-gray-900 focus:ring-2 focus:ring-klein-blue/20 focus:border-klein-blue transition-all outline-none placeholder:text-gray-400"
                placeholder="Your username"
              />
            </div>
          </div>

          <div className="space-y-1.5">
            <div className="flex justify-between items-center">
              <label className="text-sm font-medium text-gray-700">Password</label>
              {isLogin && (
                <button type="button" className="text-sm text-klein-blue hover:text-blue-800 transition-colors font-medium">
                  Forgot?
                </button>
              )}
            </div>
            <div className="relative">
              <Lock className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4.5 h-4.5 text-gray-400" />
              <input 
                type="password" 
                className="w-full bg-white border border-gray-200 rounded-xl p-3 pl-11 text-gray-900 focus:ring-2 focus:ring-klein-blue/20 focus:border-klein-blue transition-all outline-none placeholder:text-gray-400"
                placeholder="••••••••"
              />
            </div>
          </div>

          <motion.button
            whileHover={{ scale: 1.01 }}
            whileTap={{ scale: 0.99 }}
            className="w-full bg-klein-blue hover:bg-blue-800 text-white font-semibold py-3.5 rounded-xl transition-all flex items-center justify-center gap-2 shadow-lg shadow-klein-blue/20"
          >
            {isLogin ? 'Sign In' : 'Create Account'}
            <ArrowRight className="w-4 h-4" />
          </motion.button>
        </form>

        <div className="mt-8 text-center text-sm">
          <span className="text-gray-500">
            {isLogin ? "New to Minor?" : "Already have an account?"}
          </span>
          <button 
            onClick={toggleMode}
            className="text-klein-blue hover:underline ml-1.5 font-semibold"
          >
            {isLogin ? 'Join now' : 'Log in instead'}
          </button>
        </div>

        <div className="relative my-10">
          <div className="absolute inset-0 flex items-center">
            <div className="w-full border-t border-gray-100"></div>
          </div>
          <div className="relative flex justify-center text-xs uppercase tracking-widest">
            <span className="bg-white px-4 text-gray-400 font-medium">Platform Login</span>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <button className="flex items-center justify-center gap-2.5 py-3 bg-white hover:bg-gray-50 border border-gray-200 rounded-xl transition-all text-sm font-semibold text-gray-700 hover:border-gray-300">
            <Cpu className="w-4.5 h-4.5 text-gray-500" /> GitHub
          </button>
          <button className="flex items-center justify-center gap-2.5 py-3 bg-white hover:bg-gray-50 border border-gray-200 rounded-xl transition-all text-sm font-semibold text-gray-700 hover:border-gray-300">
            <Globe className="w-4.5 h-4.5 text-gray-500" /> Google
          </button>
        </div>
      </motion.div>
    </div>
  );
};

export default AuthPage;
