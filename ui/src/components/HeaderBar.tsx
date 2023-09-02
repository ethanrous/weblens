
import * as React from 'react';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import IconButton from '@mui/material/IconButton';
import MenuIcon from '@mui/icons-material/Menu';
import UploadIcon from '@mui/icons-material/Upload';
import SyncIcon from '@mui/icons-material/Sync';
import Box from '@mui/material/Box';
import { useNavigate } from "react-router-dom";


const HeaderBar = (props) => {
    let navigate = useNavigate();
    const routeChange = () => {
        let path = `/upload`;
        navigate(path);
    }

    const syncDatabase = () => {
        fetch(
            '/api/scan',
            {
                method: 'POST',
            }
        )
    }

    return (
        <Box sx={{ flexGrow: 1 }}>
            <AppBar
                position="static"
                color='transparent'
            >
                <Toolbar>
                    <IconButton edge="start" color="inherit" aria-label="menu" sx={{ mr: 2 }}>
                        <MenuIcon />
                    </IconButton>
                    <Typography variant="h6" color="inherit" component="div" sx={{ flexGrow: 1 }}>
                        Photos
                    </Typography>
                    <IconButton onClick={syncDatabase} edge="end" color="inherit" aria-label="upload" sx={{ mr: 2 }}>
                        <SyncIcon />
                    </IconButton>
                    <IconButton onClick={routeChange} edge="end" color="inherit" aria-label="upload" sx={{ mr: 2 }}>
                        <UploadIcon />
                    </IconButton>
                </Toolbar>
            </AppBar>
        </Box>
    );
}

export default HeaderBar