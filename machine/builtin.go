package machine

// BuiltinLibrary returns the default library of known machines.
func BuiltinLibrary() *Library {
	return &Library{Machines: []Machine{
		// --- Diode lasers ---
		{
			Name: "FoxAlien Reizer", Brand: "FoxAlien", Type: DiodeLaser,
			WorkAreaX: 300, WorkAreaY: 220,
			MaxFeedX: 20000, MaxFeedY: 20000,
			AccelX: 400, AccelY: 400,
			Origin: FrontLeft, LaserMode: true, MaxPower: 1000,
		},
		{
			Name: "Ortur Laser Master 2 S2", Brand: "Ortur", Type: DiodeLaser,
			WorkAreaX: 390, WorkAreaY: 410,
			MaxFeedX: 5000, MaxFeedY: 5000,
			AccelX: 500, AccelY: 500,
			Origin: FrontLeft, LaserMode: true, MaxPower: 1000,
		},
		{
			Name: "Ortur Laser Master 3", Brand: "Ortur", Type: DiodeLaser,
			WorkAreaX: 400, WorkAreaY: 400,
			MaxFeedX: 20000, MaxFeedY: 20000,
			AccelX: 2500, AccelY: 2500,
			Origin: FrontLeft, LaserMode: true, MaxPower: 1000,
		},
		{
			Name: "Sculpfun S9", Brand: "Sculpfun", Type: DiodeLaser,
			WorkAreaX: 410, WorkAreaY: 420,
			MaxFeedX: 6000, MaxFeedY: 6000,
			AccelX: 1000, AccelY: 1000,
			Origin: FrontLeft, LaserMode: true, MaxPower: 1000,
		},
		{
			Name: "Sculpfun S10", Brand: "Sculpfun", Type: DiodeLaser,
			WorkAreaX: 410, WorkAreaY: 400,
			MaxFeedX: 6000, MaxFeedY: 6000,
			AccelX: 1000, AccelY: 1000,
			Origin: FrontLeft, LaserMode: true, MaxPower: 1000,
		},
		{
			Name: "Sculpfun S30 Pro Max", Brand: "Sculpfun", Type: DiodeLaser,
			WorkAreaX: 370, WorkAreaY: 360,
			MaxFeedX: 6000, MaxFeedY: 6000,
			AccelX: 1000, AccelY: 1000,
			Origin: FrontLeft, LaserMode: true, MaxPower: 1000,
			GasAssist: AirAssist,
		},
		{
			Name: "Atomstack A5 Pro", Brand: "Atomstack", Type: DiodeLaser,
			WorkAreaX: 410, WorkAreaY: 400,
			MaxFeedX: 6000, MaxFeedY: 6000,
			AccelX: 1000, AccelY: 1000,
			Origin: FrontLeft, LaserMode: true, MaxPower: 1000,
		},
		{
			Name: "Atomstack X7 Pro", Brand: "Atomstack", Type: DiodeLaser,
			WorkAreaX: 410, WorkAreaY: 400,
			MaxFeedX: 6000, MaxFeedY: 6000,
			AccelX: 1000, AccelY: 1000,
			Origin: FrontLeft, LaserMode: true, MaxPower: 1000,
		},
		{
			Name: "Atomstack X20 Pro", Brand: "Atomstack", Type: DiodeLaser,
			WorkAreaX: 400, WorkAreaY: 400,
			MaxFeedX: 6000, MaxFeedY: 6000,
			AccelX: 1000, AccelY: 1000,
			Origin: FrontLeft, LaserMode: true, MaxPower: 1000,
		},
		{
			Name: "Atomstack X40 Pro", Brand: "Atomstack", Type: DiodeLaser,
			WorkAreaX: 400, WorkAreaY: 400,
			MaxFeedX: 6000, MaxFeedY: 6000,
			AccelX: 1000, AccelY: 1000,
			Origin: FrontLeft, LaserMode: true, MaxPower: 1000,
			GasAssist: AirAssist,
		},
		{
			Name: "xTool D1", Brand: "xTool", Type: DiodeLaser,
			WorkAreaX: 432, WorkAreaY: 406,
			MaxFeedX: 24000, MaxFeedY: 24000,
			AccelX: 3000, AccelY: 3000,
			Origin: FrontLeft, LaserMode: true, MaxPower: 1000,
		},
		{
			Name: "xTool D1 Pro 10W", Brand: "xTool", Type: DiodeLaser,
			WorkAreaX: 430, WorkAreaY: 400,
			MaxFeedX: 24000, MaxFeedY: 24000,
			AccelX: 3000, AccelY: 3000,
			Origin: FrontLeft, LaserMode: true, MaxPower: 1000,
		},
		{
			Name: "xTool D1 Pro 20W", Brand: "xTool", Type: DiodeLaser,
			WorkAreaX: 390, WorkAreaY: 400,
			MaxFeedX: 24000, MaxFeedY: 24000,
			AccelX: 3000, AccelY: 3000,
			Origin: FrontLeft, LaserMode: true, MaxPower: 1000,
		},
		{
			Name: "TwoTrees TTS-55", Brand: "TwoTrees", Type: DiodeLaser,
			WorkAreaX: 300, WorkAreaY: 300,
			MaxFeedX: 10000, MaxFeedY: 10000,
			AccelX: 1000, AccelY: 1000,
			Origin: FrontLeft, LaserMode: true, MaxPower: 1000,
		},
		{
			Name: "TwoTrees TS2", Brand: "TwoTrees", Type: DiodeLaser,
			WorkAreaX: 450, WorkAreaY: 450,
			MaxFeedX: 10000, MaxFeedY: 10000,
			AccelX: 1000, AccelY: 1000,
			Origin: FrontLeft, LaserMode: true, MaxPower: 1000,
		},

		// --- CO₂ lasers ---
		{
			Name: "K40 (GRBL)", Brand: "Generic", Type: CO2Laser,
			WorkAreaX: 300, WorkAreaY: 200,
			MaxFeedX: 30000, MaxFeedY: 30000,
			AccelX: 500, AccelY: 500,
			Origin: RearLeft, LaserMode: true, MaxPower: 1000,
		},
		{
			Name: "OMTech K40+ (GRBL)", Brand: "OMTech", Type: CO2Laser,
			WorkAreaX: 300, WorkAreaY: 200,
			MaxFeedX: 18000, MaxFeedY: 18000,
			AccelX: 500, AccelY: 500,
			Origin: RearLeft, LaserMode: true, MaxPower: 1000,
			GasAssist: AirAssist,
		},

		// --- CNC routers (3-axis, laser optional) ---
		{
			Name: "FoxAlien Masuter Pro", Brand: "FoxAlien", Type: CNCRouter,
			WorkAreaX: 400, WorkAreaY: 380, WorkAreaZ: 55,
			MaxFeedX: 2000, MaxFeedY: 2000, MaxFeedZ: 1200,
			AccelX: 300, AccelY: 300, AccelZ: 30,
			Origin: FrontLeft, LaserMode: false, MaxPower: 10000,
		},
		{
			Name: "FoxAlien 4040-XE", Brand: "FoxAlien", Type: CNCRouter,
			WorkAreaX: 400, WorkAreaY: 400, WorkAreaZ: 65,
			MaxFeedX: 2000, MaxFeedY: 2000, MaxFeedZ: 1200,
			AccelX: 300, AccelY: 300, AccelZ: 30,
			Origin: FrontLeft, LaserMode: false, MaxPower: 10000,
		},
		{
			Name: "FoxAlien XE-Pro", Brand: "FoxAlien", Type: CNCRouter,
			WorkAreaX: 400, WorkAreaY: 400, WorkAreaZ: 65,
			MaxFeedX: 2000, MaxFeedY: 2000, MaxFeedZ: 1200,
			AccelX: 300, AccelY: 300, AccelZ: 30,
			Origin: FrontLeft, LaserMode: false, MaxPower: 1000,
			GasAssist: CO2Assist,
		},
		{
			Name: "SainSmart 3018 Pro", Brand: "SainSmart", Type: CNCRouter,
			WorkAreaX: 300, WorkAreaY: 180, WorkAreaZ: 45,
			MaxFeedX: 1000, MaxFeedY: 1000, MaxFeedZ: 800,
			AccelX: 10, AccelY: 10, AccelZ: 10,
			Origin: FrontLeft, LaserMode: false, MaxPower: 1000,
		},
		{
			Name: "SainSmart 3020-PRO Max", Brand: "SainSmart", Type: CNCRouter,
			WorkAreaX: 300, WorkAreaY: 200, WorkAreaZ: 72,
			MaxFeedX: 1000, MaxFeedY: 1000, MaxFeedZ: 800,
			AccelX: 10, AccelY: 10, AccelZ: 10,
			Origin: FrontLeft, LaserMode: false, MaxPower: 1000,
		},

		// --- Fiber laser (industrial, O₂ assist) ---
		{
			Name: "Fiber Laser 1000W (Generic)", Brand: "Generic", Type: FiberLaser,
			WorkAreaX: 1500, WorkAreaY: 3000,
			MaxFeedX: 60000, MaxFeedY: 60000,
			AccelX: 5000, AccelY: 5000,
			Origin: FrontLeft, LaserMode: true, MaxPower: 1000,
			GasAssist: O2Assist,
		},
	}}
}
