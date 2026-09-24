/*eslint-disable block-scoped-var, id-length, no-control-regex, no-magic-numbers, no-prototype-builtins, no-redeclare, no-shadow, no-var, sort-vars*/
import * as $protobuf from "protobufjs/minimal";

const $Reader = $protobuf.Reader, $Writer = $protobuf.Writer, $util = $protobuf.util;

const $root = $protobuf.roots.roots || ($protobuf.roots.roots = {});

export const greyvarproto = $root.greyvarproto = (() => {

    const greyvarproto = {};

    greyvarproto.ClientRequests = (function() {

        function ClientRequests(p) {
            if (p)
                for (var ks = Object.keys(p), i = 0; i < ks.length; ++i)
                    if (p[ks[i]] != null)
                        this[ks[i]] = p[ks[i]];
        }

        ClientRequests.prototype.registrationRequest = null;
        ClientRequests.prototype.moveRequest = null;

        ClientRequests.create = function create(properties) {
            return new ClientRequests(properties);
        };

        ClientRequests.encode = function encode(m, w) {
            if (!w)
                w = $Writer.create();
            if (m.registrationRequest != null && Object.hasOwnProperty.call(m, "registrationRequest"))
                $root.greyvarproto.RegistrationRequest.encode(m.registrationRequest, w.uint32(10).fork()).ldelim();
            if (m.moveRequest != null && Object.hasOwnProperty.call(m, "moveRequest"))
                $root.greyvarproto.MoveRequest.encode(m.moveRequest, w.uint32(18).fork()).ldelim();
            return w;
        };

        ClientRequests.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        ClientRequests.decode = function decode(r, l) {
            if (!(r instanceof $Reader))
                r = $Reader.create(r);
            var c = l === undefined ? r.len : r.pos + l, m = new $root.greyvarproto.ClientRequests();
            while (r.pos < c) {
                var t = r.uint32();
                switch (t >>> 3) {
                case 1:
                    m.registrationRequest = $root.greyvarproto.RegistrationRequest.decode(r, r.uint32());
                    break;
                case 2:
                    m.moveRequest = $root.greyvarproto.MoveRequest.decode(r, r.uint32());
                    break;
                default:
                    r.skipType(t & 7);
                    break;
                }
            }
            return m;
        };

        ClientRequests.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        ClientRequests.verify = function verify(m) {
            if (typeof m !== "object" || m === null)
                return "object expected";
            if (m.registrationRequest != null && m.hasOwnProperty("registrationRequest")) {
                {
                    var e = $root.greyvarproto.RegistrationRequest.verify(m.registrationRequest);
                    if (e)
                        return "registrationRequest." + e;
                }
            }
            if (m.moveRequest != null && m.hasOwnProperty("moveRequest")) {
                {
                    var e = $root.greyvarproto.MoveRequest.verify(m.moveRequest);
                    if (e)
                        return "moveRequest." + e;
                }
            }
            return null;
        };

        ClientRequests.fromObject = function fromObject(d) {
            if (d instanceof $root.greyvarproto.ClientRequests)
                return d;
            var m = new $root.greyvarproto.ClientRequests();
            if (d.registrationRequest != null) {
                if (typeof d.registrationRequest !== "object")
                    throw TypeError(".greyvarproto.ClientRequests.registrationRequest: object expected");
                m.registrationRequest = $root.greyvarproto.RegistrationRequest.fromObject(d.registrationRequest);
            }
            if (d.moveRequest != null) {
                if (typeof d.moveRequest !== "object")
                    throw TypeError(".greyvarproto.ClientRequests.moveRequest: object expected");
                m.moveRequest = $root.greyvarproto.MoveRequest.fromObject(d.moveRequest);
            }
            return m;
        };

        ClientRequests.toObject = function toObject(m, o) {
            if (!o)
                o = {};
            var d = {};
            if (o.defaults) {
                d.registrationRequest = null;
                d.moveRequest = null;
            }
            if (m.registrationRequest != null && m.hasOwnProperty("registrationRequest")) {
                d.registrationRequest = $root.greyvarproto.RegistrationRequest.toObject(m.registrationRequest, o);
            }
            if (m.moveRequest != null && m.hasOwnProperty("moveRequest")) {
                d.moveRequest = $root.greyvarproto.MoveRequest.toObject(m.moveRequest, o);
            }
            return d;
        };

        ClientRequests.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        return ClientRequests;
    })();

    greyvarproto.MoveRequest = (function() {

        function MoveRequest(p) {
            if (p)
                for (var ks = Object.keys(p), i = 0; i < ks.length; ++i)
                    if (p[ks[i]] != null)
                        this[ks[i]] = p[ks[i]];
        }

        MoveRequest.prototype.playerId = 0;
        MoveRequest.prototype.x = 0;
        MoveRequest.prototype.y = 0;

        MoveRequest.create = function create(properties) {
            return new MoveRequest(properties);
        };

        MoveRequest.encode = function encode(m, w) {
            if (!w)
                w = $Writer.create();
            if (m.playerId != null && Object.hasOwnProperty.call(m, "playerId"))
                w.uint32(8).int32(m.playerId);
            if (m.x != null && Object.hasOwnProperty.call(m, "x"))
                w.uint32(16).int32(m.x);
            if (m.y != null && Object.hasOwnProperty.call(m, "y"))
                w.uint32(24).int32(m.y);
            return w;
        };

        MoveRequest.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        MoveRequest.decode = function decode(r, l) {
            if (!(r instanceof $Reader))
                r = $Reader.create(r);
            var c = l === undefined ? r.len : r.pos + l, m = new $root.greyvarproto.MoveRequest();
            while (r.pos < c) {
                var t = r.uint32();
                switch (t >>> 3) {
                case 1:
                    m.playerId = r.int32();
                    break;
                case 2:
                    m.x = r.int32();
                    break;
                case 3:
                    m.y = r.int32();
                    break;
                default:
                    r.skipType(t & 7);
                    break;
                }
            }
            return m;
        };

        MoveRequest.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        MoveRequest.verify = function verify(m) {
            if (typeof m !== "object" || m === null)
                return "object expected";
            if (m.playerId != null && m.hasOwnProperty("playerId")) {
                if (!$util.isInteger(m.playerId))
                    return "playerId: integer expected";
            }
            if (m.x != null && m.hasOwnProperty("x")) {
                if (!$util.isInteger(m.x))
                    return "x: integer expected";
            }
            if (m.y != null && m.hasOwnProperty("y")) {
                if (!$util.isInteger(m.y))
                    return "y: integer expected";
            }
            return null;
        };

        MoveRequest.fromObject = function fromObject(d) {
            if (d instanceof $root.greyvarproto.MoveRequest)
                return d;
            var m = new $root.greyvarproto.MoveRequest();
            if (d.playerId != null) {
                m.playerId = d.playerId | 0;
            }
            if (d.x != null) {
                m.x = d.x | 0;
            }
            if (d.y != null) {
                m.y = d.y | 0;
            }
            return m;
        };

        MoveRequest.toObject = function toObject(m, o) {
            if (!o)
                o = {};
            var d = {};
            if (o.defaults) {
                d.playerId = 0;
                d.x = 0;
                d.y = 0;
            }
            if (m.playerId != null && m.hasOwnProperty("playerId")) {
                d.playerId = m.playerId;
            }
            if (m.x != null && m.hasOwnProperty("x")) {
                d.x = m.x;
            }
            if (m.y != null && m.hasOwnProperty("y")) {
                d.y = m.y;
            }
            return d;
        };

        MoveRequest.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        return MoveRequest;
    })();

    greyvarproto.ConnectionRequest = (function() {

        function ConnectionRequest(p) {
            if (p)
                for (var ks = Object.keys(p), i = 0; i < ks.length; ++i)
                    if (p[ks[i]] != null)
                        this[ks[i]] = p[ks[i]];
        }

        ConnectionRequest.create = function create(properties) {
            return new ConnectionRequest(properties);
        };

        ConnectionRequest.encode = function encode(m, w) {
            if (!w)
                w = $Writer.create();
            return w;
        };

        ConnectionRequest.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        ConnectionRequest.decode = function decode(r, l) {
            if (!(r instanceof $Reader))
                r = $Reader.create(r);
            var c = l === undefined ? r.len : r.pos + l, m = new $root.greyvarproto.ConnectionRequest();
            while (r.pos < c) {
                var t = r.uint32();
                switch (t >>> 3) {
                default:
                    r.skipType(t & 7);
                    break;
                }
            }
            return m;
        };

        ConnectionRequest.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        ConnectionRequest.verify = function verify(m) {
            if (typeof m !== "object" || m === null)
                return "object expected";
            return null;
        };

        ConnectionRequest.fromObject = function fromObject(d) {
            if (d instanceof $root.greyvarproto.ConnectionRequest)
                return d;
            return new $root.greyvarproto.ConnectionRequest();
        };

        ConnectionRequest.toObject = function toObject() {
            return {};
        };

        ConnectionRequest.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        return ConnectionRequest;
    })();

    greyvarproto.RegistrationRequest = (function() {

        function RegistrationRequest(p) {
            if (p)
                for (var ks = Object.keys(p), i = 0; i < ks.length; ++i)
                    if (p[ks[i]] != null)
                        this[ks[i]] = p[ks[i]];
        }

        RegistrationRequest.prototype.username = "";

        RegistrationRequest.create = function create(properties) {
            return new RegistrationRequest(properties);
        };

        RegistrationRequest.encode = function encode(m, w) {
            if (!w)
                w = $Writer.create();
            if (m.username != null && Object.hasOwnProperty.call(m, "username"))
                w.uint32(10).string(m.username);
            return w;
        };

        RegistrationRequest.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        RegistrationRequest.decode = function decode(r, l) {
            if (!(r instanceof $Reader))
                r = $Reader.create(r);
            var c = l === undefined ? r.len : r.pos + l, m = new $root.greyvarproto.RegistrationRequest();
            while (r.pos < c) {
                var t = r.uint32();
                switch (t >>> 3) {
                case 1:
                    m.username = r.string();
                    break;
                default:
                    r.skipType(t & 7);
                    break;
                }
            }
            return m;
        };

        RegistrationRequest.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        RegistrationRequest.verify = function verify(m) {
            if (typeof m !== "object" || m === null)
                return "object expected";
            if (m.username != null && m.hasOwnProperty("username")) {
                if (!$util.isString(m.username))
                    return "username: string expected";
            }
            return null;
        };

        RegistrationRequest.fromObject = function fromObject(d) {
            if (d instanceof $root.greyvarproto.RegistrationRequest)
                return d;
            var m = new $root.greyvarproto.RegistrationRequest();
            if (d.username != null) {
                m.username = String(d.username);
            }
            return m;
        };

        RegistrationRequest.toObject = function toObject(m, o) {
            if (!o)
                o = {};
            var d = {};
            if (o.defaults) {
                d.username = "";
            }
            if (m.username != null && m.hasOwnProperty("username")) {
                d.username = m.username;
            }
            return d;
        };

        RegistrationRequest.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        return RegistrationRequest;
    })();

    greyvarproto.ServerUpdate = (function() {

        function ServerUpdate(p) {
            this.entityPositions = [];
            this.entitySpawns = [];
            this.entityDefinitions = [];
            this.entityStateChanges = [];
            if (p)
                for (var ks = Object.keys(p), i = 0; i < ks.length; ++i)
                    if (p[ks[i]] != null)
                        this[ks[i]] = p[ks[i]];
        }

        ServerUpdate.prototype.entityPositions = $util.emptyArray;
        ServerUpdate.prototype.entitySpawns = $util.emptyArray;
        ServerUpdate.prototype.playerJoined = null;
        ServerUpdate.prototype.singleTileChange = null;
        ServerUpdate.prototype.grid = null;
        ServerUpdate.prototype.connectionResponse = null;
        ServerUpdate.prototype.entityDefinitions = $util.emptyArray;
        ServerUpdate.prototype.entityStateChanges = $util.emptyArray;

        ServerUpdate.create = function create(properties) {
            return new ServerUpdate(properties);
        };

        ServerUpdate.encode = function encode(m, w) {
            if (!w)
                w = $Writer.create();
            if (m.entityPositions != null && m.entityPositions.length) {
                for (var i = 0; i < m.entityPositions.length; ++i)
                    $root.greyvarproto.EntityPosition.encode(m.entityPositions[i], w.uint32(18).fork()).ldelim();
            }
            if (m.entitySpawns != null && m.entitySpawns.length) {
                for (var i = 0; i < m.entitySpawns.length; ++i)
                    $root.greyvarproto.EntitySpawn.encode(m.entitySpawns[i], w.uint32(26).fork()).ldelim();
            }
            if (m.playerJoined != null && Object.hasOwnProperty.call(m, "playerJoined"))
                $root.greyvarproto.PlayerJoined.encode(m.playerJoined, w.uint32(34).fork()).ldelim();
            if (m.singleTileChange != null && Object.hasOwnProperty.call(m, "singleTileChange"))
                $root.greyvarproto.Tile.encode(m.singleTileChange, w.uint32(42).fork()).ldelim();
            if (m.grid != null && Object.hasOwnProperty.call(m, "grid"))
                $root.greyvarproto.Grid.encode(m.grid, w.uint32(50).fork()).ldelim();
            if (m.connectionResponse != null && Object.hasOwnProperty.call(m, "connectionResponse"))
                $root.greyvarproto.ConnectionResponse.encode(m.connectionResponse, w.uint32(58).fork()).ldelim();
            if (m.entityDefinitions != null && m.entityDefinitions.length) {
                for (var i = 0; i < m.entityDefinitions.length; ++i)
                    $root.greyvarproto.EntityDefinition.encode(m.entityDefinitions[i], w.uint32(66).fork()).ldelim();
            }
            if (m.entityStateChanges != null && m.entityStateChanges.length) {
                for (var i = 0; i < m.entityStateChanges.length; ++i)
                    $root.greyvarproto.EntityStateChange.encode(m.entityStateChanges[i], w.uint32(74).fork()).ldelim();
            }
            return w;
        };

        ServerUpdate.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        ServerUpdate.decode = function decode(r, l) {
            if (!(r instanceof $Reader))
                r = $Reader.create(r);
            var c = l === undefined ? r.len : r.pos + l, m = new $root.greyvarproto.ServerUpdate();
            while (r.pos < c) {
                var t = r.uint32();
                switch (t >>> 3) {
                case 2:
                    if (!(m.entityPositions && m.entityPositions.length))
                        m.entityPositions = [];
                    m.entityPositions.push($root.greyvarproto.EntityPosition.decode(r, r.uint32()));
                    break;
                case 3:
                    if (!(m.entitySpawns && m.entitySpawns.length))
                        m.entitySpawns = [];
                    m.entitySpawns.push($root.greyvarproto.EntitySpawn.decode(r, r.uint32()));
                    break;
                case 4:
                    m.playerJoined = $root.greyvarproto.PlayerJoined.decode(r, r.uint32());
                    break;
                case 5:
                    m.singleTileChange = $root.greyvarproto.Tile.decode(r, r.uint32());
                    break;
                case 6:
                    m.grid = $root.greyvarproto.Grid.decode(r, r.uint32());
                    break;
                case 7:
                    m.connectionResponse = $root.greyvarproto.ConnectionResponse.decode(r, r.uint32());
                    break;
                case 8:
                    if (!(m.entityDefinitions && m.entityDefinitions.length))
                        m.entityDefinitions = [];
                    m.entityDefinitions.push($root.greyvarproto.EntityDefinition.decode(r, r.uint32()));
                    break;
                case 9:
                    if (!(m.entityStateChanges && m.entityStateChanges.length))
                        m.entityStateChanges = [];
                    m.entityStateChanges.push($root.greyvarproto.EntityStateChange.decode(r, r.uint32()));
                    break;
                default:
                    r.skipType(t & 7);
                    break;
                }
            }
            return m;
        };

        ServerUpdate.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        ServerUpdate.verify = function verify(m) {
            if (typeof m !== "object" || m === null)
                return "object expected";
            if (m.entityPositions != null && m.hasOwnProperty("entityPositions")) {
                if (!Array.isArray(m.entityPositions))
                    return "entityPositions: array expected";
                for (var i = 0; i < m.entityPositions.length; ++i) {
                    {
                        var e = $root.greyvarproto.EntityPosition.verify(m.entityPositions[i]);
                        if (e)
                            return "entityPositions." + e;
                    }
                }
            }
            if (m.entitySpawns != null && m.hasOwnProperty("entitySpawns")) {
                if (!Array.isArray(m.entitySpawns))
                    return "entitySpawns: array expected";
                for (var i = 0; i < m.entitySpawns.length; ++i) {
                    {
                        var e = $root.greyvarproto.EntitySpawn.verify(m.entitySpawns[i]);
                        if (e)
                            return "entitySpawns." + e;
                    }
                }
            }
            if (m.playerJoined != null && m.hasOwnProperty("playerJoined")) {
                {
                    var e = $root.greyvarproto.PlayerJoined.verify(m.playerJoined);
                    if (e)
                        return "playerJoined." + e;
                }
            }
            if (m.singleTileChange != null && m.hasOwnProperty("singleTileChange")) {
                {
                    var e = $root.greyvarproto.Tile.verify(m.singleTileChange);
                    if (e)
                        return "singleTileChange." + e;
                }
            }
            if (m.grid != null && m.hasOwnProperty("grid")) {
                {
                    var e = $root.greyvarproto.Grid.verify(m.grid);
                    if (e)
                        return "grid." + e;
                }
            }
            if (m.connectionResponse != null && m.hasOwnProperty("connectionResponse")) {
                {
                    var e = $root.greyvarproto.ConnectionResponse.verify(m.connectionResponse);
                    if (e)
                        return "connectionResponse." + e;
                }
            }
            if (m.entityDefinitions != null && m.hasOwnProperty("entityDefinitions")) {
                if (!Array.isArray(m.entityDefinitions))
                    return "entityDefinitions: array expected";
                for (var i = 0; i < m.entityDefinitions.length; ++i) {
                    {
                        var e = $root.greyvarproto.EntityDefinition.verify(m.entityDefinitions[i]);
                        if (e)
                            return "entityDefinitions." + e;
                    }
                }
            }
            if (m.entityStateChanges != null && m.hasOwnProperty("entityStateChanges")) {
                if (!Array.isArray(m.entityStateChanges))
                    return "entityStateChanges: array expected";
                for (var i = 0; i < m.entityStateChanges.length; ++i) {
                    {
                        var e = $root.greyvarproto.EntityStateChange.verify(m.entityStateChanges[i]);
                        if (e)
                            return "entityStateChanges." + e;
                    }
                }
            }
            return null;
        };

        ServerUpdate.fromObject = function fromObject(d) {
            if (d instanceof $root.greyvarproto.ServerUpdate)
                return d;
            var m = new $root.greyvarproto.ServerUpdate();
            if (d.entityPositions) {
                if (!Array.isArray(d.entityPositions))
                    throw TypeError(".greyvarproto.ServerUpdate.entityPositions: array expected");
                m.entityPositions = [];
                for (var i = 0; i < d.entityPositions.length; ++i) {
                    if (typeof d.entityPositions[i] !== "object")
                        throw TypeError(".greyvarproto.ServerUpdate.entityPositions: object expected");
                    m.entityPositions[i] = $root.greyvarproto.EntityPosition.fromObject(d.entityPositions[i]);
                }
            }
            if (d.entitySpawns) {
                if (!Array.isArray(d.entitySpawns))
                    throw TypeError(".greyvarproto.ServerUpdate.entitySpawns: array expected");
                m.entitySpawns = [];
                for (var i = 0; i < d.entitySpawns.length; ++i) {
                    if (typeof d.entitySpawns[i] !== "object")
                        throw TypeError(".greyvarproto.ServerUpdate.entitySpawns: object expected");
                    m.entitySpawns[i] = $root.greyvarproto.EntitySpawn.fromObject(d.entitySpawns[i]);
                }
            }
            if (d.playerJoined != null) {
                if (typeof d.playerJoined !== "object")
                    throw TypeError(".greyvarproto.ServerUpdate.playerJoined: object expected");
                m.playerJoined = $root.greyvarproto.PlayerJoined.fromObject(d.playerJoined);
            }
            if (d.singleTileChange != null) {
                if (typeof d.singleTileChange !== "object")
                    throw TypeError(".greyvarproto.ServerUpdate.singleTileChange: object expected");
                m.singleTileChange = $root.greyvarproto.Tile.fromObject(d.singleTileChange);
            }
            if (d.grid != null) {
                if (typeof d.grid !== "object")
                    throw TypeError(".greyvarproto.ServerUpdate.grid: object expected");
                m.grid = $root.greyvarproto.Grid.fromObject(d.grid);
            }
            if (d.connectionResponse != null) {
                if (typeof d.connectionResponse !== "object")
                    throw TypeError(".greyvarproto.ServerUpdate.connectionResponse: object expected");
                m.connectionResponse = $root.greyvarproto.ConnectionResponse.fromObject(d.connectionResponse);
            }
            if (d.entityDefinitions) {
                if (!Array.isArray(d.entityDefinitions))
                    throw TypeError(".greyvarproto.ServerUpdate.entityDefinitions: array expected");
                m.entityDefinitions = [];
                for (var i = 0; i < d.entityDefinitions.length; ++i) {
                    if (typeof d.entityDefinitions[i] !== "object")
                        throw TypeError(".greyvarproto.ServerUpdate.entityDefinitions: object expected");
                    m.entityDefinitions[i] = $root.greyvarproto.EntityDefinition.fromObject(d.entityDefinitions[i]);
                }
            }
            if (d.entityStateChanges) {
                if (!Array.isArray(d.entityStateChanges))
                    throw TypeError(".greyvarproto.ServerUpdate.entityStateChanges: array expected");
                m.entityStateChanges = [];
                for (var i = 0; i < d.entityStateChanges.length; ++i) {
                    if (typeof d.entityStateChanges[i] !== "object")
                        throw TypeError(".greyvarproto.ServerUpdate.entityStateChanges: object expected");
                    m.entityStateChanges[i] = $root.greyvarproto.EntityStateChange.fromObject(d.entityStateChanges[i]);
                }
            }
            return m;
        };

        ServerUpdate.toObject = function toObject(m, o) {
            if (!o)
                o = {};
            var d = {};
            if (o.arrays || o.defaults) {
                d.entityPositions = [];
                d.entitySpawns = [];
                d.entityDefinitions = [];
                d.entityStateChanges = [];
            }
            if (o.defaults) {
                d.playerJoined = null;
                d.singleTileChange = null;
                d.grid = null;
                d.connectionResponse = null;
            }
            if (m.entityPositions && m.entityPositions.length) {
                d.entityPositions = [];
                for (var j = 0; j < m.entityPositions.length; ++j) {
                    d.entityPositions[j] = $root.greyvarproto.EntityPosition.toObject(m.entityPositions[j], o);
                }
            }
            if (m.entitySpawns && m.entitySpawns.length) {
                d.entitySpawns = [];
                for (var j = 0; j < m.entitySpawns.length; ++j) {
                    d.entitySpawns[j] = $root.greyvarproto.EntitySpawn.toObject(m.entitySpawns[j], o);
                }
            }
            if (m.playerJoined != null && m.hasOwnProperty("playerJoined")) {
                d.playerJoined = $root.greyvarproto.PlayerJoined.toObject(m.playerJoined, o);
            }
            if (m.singleTileChange != null && m.hasOwnProperty("singleTileChange")) {
                d.singleTileChange = $root.greyvarproto.Tile.toObject(m.singleTileChange, o);
            }
            if (m.grid != null && m.hasOwnProperty("grid")) {
                d.grid = $root.greyvarproto.Grid.toObject(m.grid, o);
            }
            if (m.connectionResponse != null && m.hasOwnProperty("connectionResponse")) {
                d.connectionResponse = $root.greyvarproto.ConnectionResponse.toObject(m.connectionResponse, o);
            }
            if (m.entityDefinitions && m.entityDefinitions.length) {
                d.entityDefinitions = [];
                for (var j = 0; j < m.entityDefinitions.length; ++j) {
                    d.entityDefinitions[j] = $root.greyvarproto.EntityDefinition.toObject(m.entityDefinitions[j], o);
                }
            }
            if (m.entityStateChanges && m.entityStateChanges.length) {
                d.entityStateChanges = [];
                for (var j = 0; j < m.entityStateChanges.length; ++j) {
                    d.entityStateChanges[j] = $root.greyvarproto.EntityStateChange.toObject(m.entityStateChanges[j], o);
                }
            }
            return d;
        };

        ServerUpdate.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        return ServerUpdate;
    })();

    greyvarproto.EntityPosition = (function() {

        function EntityPosition(p) {
            if (p)
                for (var ks = Object.keys(p), i = 0; i < ks.length; ++i)
                    if (p[ks[i]] != null)
                        this[ks[i]] = p[ks[i]];
        }

        EntityPosition.prototype.entityId = $util.Long ? $util.Long.fromBits(0,0,false) : 0;
        EntityPosition.prototype.x = 0;
        EntityPosition.prototype.y = 0;

        EntityPosition.create = function create(properties) {
            return new EntityPosition(properties);
        };

        EntityPosition.encode = function encode(m, w) {
            if (!w)
                w = $Writer.create();
            if (m.entityId != null && Object.hasOwnProperty.call(m, "entityId"))
                w.uint32(8).int64(m.entityId);
            if (m.x != null && Object.hasOwnProperty.call(m, "x"))
                w.uint32(16).int32(m.x);
            if (m.y != null && Object.hasOwnProperty.call(m, "y"))
                w.uint32(24).int32(m.y);
            return w;
        };

        EntityPosition.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        EntityPosition.decode = function decode(r, l) {
            if (!(r instanceof $Reader))
                r = $Reader.create(r);
            var c = l === undefined ? r.len : r.pos + l, m = new $root.greyvarproto.EntityPosition();
            while (r.pos < c) {
                var t = r.uint32();
                switch (t >>> 3) {
                case 1:
                    m.entityId = r.int64();
                    break;
                case 2:
                    m.x = r.int32();
                    break;
                case 3:
                    m.y = r.int32();
                    break;
                default:
                    r.skipType(t & 7);
                    break;
                }
            }
            return m;
        };

        EntityPosition.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        EntityPosition.verify = function verify(m) {
            if (typeof m !== "object" || m === null)
                return "object expected";
            if (m.entityId != null && m.hasOwnProperty("entityId")) {
                if (!$util.isInteger(m.entityId) && !(m.entityId && $util.isInteger(m.entityId.low) && $util.isInteger(m.entityId.high)))
                    return "entityId: integer|Long expected";
            }
            if (m.x != null && m.hasOwnProperty("x")) {
                if (!$util.isInteger(m.x))
                    return "x: integer expected";
            }
            if (m.y != null && m.hasOwnProperty("y")) {
                if (!$util.isInteger(m.y))
                    return "y: integer expected";
            }
            return null;
        };

        EntityPosition.fromObject = function fromObject(d) {
            if (d instanceof $root.greyvarproto.EntityPosition)
                return d;
            var m = new $root.greyvarproto.EntityPosition();
            if (d.entityId != null) {
                if ($util.Long)
                    (m.entityId = $util.Long.fromValue(d.entityId)).unsigned = false;
                else if (typeof d.entityId === "string")
                    m.entityId = parseInt(d.entityId, 10);
                else if (typeof d.entityId === "number")
                    m.entityId = d.entityId;
                else if (typeof d.entityId === "object")
                    m.entityId = new $util.LongBits(d.entityId.low >>> 0, d.entityId.high >>> 0).toNumber();
            }
            if (d.x != null) {
                m.x = d.x | 0;
            }
            if (d.y != null) {
                m.y = d.y | 0;
            }
            return m;
        };

        EntityPosition.toObject = function toObject(m, o) {
            if (!o)
                o = {};
            var d = {};
            if (o.defaults) {
                if ($util.Long) {
                    var n = new $util.Long(0, 0, false);
                    d.entityId = o.longs === String ? n.toString() : o.longs === Number ? n.toNumber() : n;
                } else
                    d.entityId = o.longs === String ? "0" : 0;
                d.x = 0;
                d.y = 0;
            }
            if (m.entityId != null && m.hasOwnProperty("entityId")) {
                if (typeof m.entityId === "number")
                    d.entityId = o.longs === String ? String(m.entityId) : m.entityId;
                else
                    d.entityId = o.longs === String ? $util.Long.prototype.toString.call(m.entityId) : o.longs === Number ? new $util.LongBits(m.entityId.low >>> 0, m.entityId.high >>> 0).toNumber() : m.entityId;
            }
            if (m.x != null && m.hasOwnProperty("x")) {
                d.x = m.x;
            }
            if (m.y != null && m.hasOwnProperty("y")) {
                d.y = m.y;
            }
            return d;
        };

        EntityPosition.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        return EntityPosition;
    })();

    greyvarproto.EntityStateChange = (function() {

        function EntityStateChange(p) {
            if (p)
                for (var ks = Object.keys(p), i = 0; i < ks.length; ++i)
                    if (p[ks[i]] != null)
                        this[ks[i]] = p[ks[i]];
        }

        EntityStateChange.prototype.entityId = $util.Long ? $util.Long.fromBits(0,0,false) : 0;
        EntityStateChange.prototype.newState = "";

        EntityStateChange.create = function create(properties) {
            return new EntityStateChange(properties);
        };

        EntityStateChange.encode = function encode(m, w) {
            if (!w)
                w = $Writer.create();
            if (m.entityId != null && Object.hasOwnProperty.call(m, "entityId"))
                w.uint32(8).int64(m.entityId);
            if (m.newState != null && Object.hasOwnProperty.call(m, "newState"))
                w.uint32(18).string(m.newState);
            return w;
        };

        EntityStateChange.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        EntityStateChange.decode = function decode(r, l) {
            if (!(r instanceof $Reader))
                r = $Reader.create(r);
            var c = l === undefined ? r.len : r.pos + l, m = new $root.greyvarproto.EntityStateChange();
            while (r.pos < c) {
                var t = r.uint32();
                switch (t >>> 3) {
                case 1:
                    m.entityId = r.int64();
                    break;
                case 2:
                    m.newState = r.string();
                    break;
                default:
                    r.skipType(t & 7);
                    break;
                }
            }
            return m;
        };

        EntityStateChange.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        EntityStateChange.verify = function verify(m) {
            if (typeof m !== "object" || m === null)
                return "object expected";
            if (m.entityId != null && m.hasOwnProperty("entityId")) {
                if (!$util.isInteger(m.entityId) && !(m.entityId && $util.isInteger(m.entityId.low) && $util.isInteger(m.entityId.high)))
                    return "entityId: integer|Long expected";
            }
            if (m.newState != null && m.hasOwnProperty("newState")) {
                if (!$util.isString(m.newState))
                    return "newState: string expected";
            }
            return null;
        };

        EntityStateChange.fromObject = function fromObject(d) {
            if (d instanceof $root.greyvarproto.EntityStateChange)
                return d;
            var m = new $root.greyvarproto.EntityStateChange();
            if (d.entityId != null) {
                if ($util.Long)
                    (m.entityId = $util.Long.fromValue(d.entityId)).unsigned = false;
                else if (typeof d.entityId === "string")
                    m.entityId = parseInt(d.entityId, 10);
                else if (typeof d.entityId === "number")
                    m.entityId = d.entityId;
                else if (typeof d.entityId === "object")
                    m.entityId = new $util.LongBits(d.entityId.low >>> 0, d.entityId.high >>> 0).toNumber();
            }
            if (d.newState != null) {
                m.newState = String(d.newState);
            }
            return m;
        };

        EntityStateChange.toObject = function toObject(m, o) {
            if (!o)
                o = {};
            var d = {};
            if (o.defaults) {
                if ($util.Long) {
                    var n = new $util.Long(0, 0, false);
                    d.entityId = o.longs === String ? n.toString() : o.longs === Number ? n.toNumber() : n;
                } else
                    d.entityId = o.longs === String ? "0" : 0;
                d.newState = "";
            }
            if (m.entityId != null && m.hasOwnProperty("entityId")) {
                if (typeof m.entityId === "number")
                    d.entityId = o.longs === String ? String(m.entityId) : m.entityId;
                else
                    d.entityId = o.longs === String ? $util.Long.prototype.toString.call(m.entityId) : o.longs === Number ? new $util.LongBits(m.entityId.low >>> 0, m.entityId.high >>> 0).toNumber() : m.entityId;
            }
            if (m.newState != null && m.hasOwnProperty("newState")) {
                d.newState = m.newState;
            }
            return d;
        };

        EntityStateChange.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        return EntityStateChange;
    })();

    greyvarproto.EntitySpawn = (function() {

        function EntitySpawn(p) {
            if (p)
                for (var ks = Object.keys(p), i = 0; i < ks.length; ++i)
                    if (p[ks[i]] != null)
                        this[ks[i]] = p[ks[i]];
        }

        EntitySpawn.prototype.entityId = $util.Long ? $util.Long.fromBits(0,0,false) : 0;
        EntitySpawn.prototype.x = 0;
        EntitySpawn.prototype.y = 0;
        EntitySpawn.prototype.w = 0;
        EntitySpawn.prototype.h = 0;
        EntitySpawn.prototype.primaryColor = 0;
        EntitySpawn.prototype.definition = "";
        EntitySpawn.prototype.initialState = "";

        EntitySpawn.create = function create(properties) {
            return new EntitySpawn(properties);
        };

        EntitySpawn.encode = function encode(m, w) {
            if (!w)
                w = $Writer.create();
            if (m.entityId != null && Object.hasOwnProperty.call(m, "entityId"))
                w.uint32(8).int64(m.entityId);
            if (m.x != null && Object.hasOwnProperty.call(m, "x"))
                w.uint32(16).int32(m.x);
            if (m.y != null && Object.hasOwnProperty.call(m, "y"))
                w.uint32(24).int32(m.y);
            if (m.w != null && Object.hasOwnProperty.call(m, "w"))
                w.uint32(32).int32(m.w);
            if (m.h != null && Object.hasOwnProperty.call(m, "h"))
                w.uint32(40).int32(m.h);
            if (m.primaryColor != null && Object.hasOwnProperty.call(m, "primaryColor"))
                w.uint32(56).uint32(m.primaryColor);
            if (m.definition != null && Object.hasOwnProperty.call(m, "definition"))
                w.uint32(74).string(m.definition);
            if (m.initialState != null && Object.hasOwnProperty.call(m, "initialState"))
                w.uint32(82).string(m.initialState);
            return w;
        };

        EntitySpawn.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        EntitySpawn.decode = function decode(r, l) {
            if (!(r instanceof $Reader))
                r = $Reader.create(r);
            var c = l === undefined ? r.len : r.pos + l, m = new $root.greyvarproto.EntitySpawn();
            while (r.pos < c) {
                var t = r.uint32();
                switch (t >>> 3) {
                case 1:
                    m.entityId = r.int64();
                    break;
                case 2:
                    m.x = r.int32();
                    break;
                case 3:
                    m.y = r.int32();
                    break;
                case 4:
                    m.w = r.int32();
                    break;
                case 5:
                    m.h = r.int32();
                    break;
                case 7:
                    m.primaryColor = r.uint32();
                    break;
                case 9:
                    m.definition = r.string();
                    break;
                case 10:
                    m.initialState = r.string();
                    break;
                default:
                    r.skipType(t & 7);
                    break;
                }
            }
            return m;
        };

        EntitySpawn.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        EntitySpawn.verify = function verify(m) {
            if (typeof m !== "object" || m === null)
                return "object expected";
            if (m.entityId != null && m.hasOwnProperty("entityId")) {
                if (!$util.isInteger(m.entityId) && !(m.entityId && $util.isInteger(m.entityId.low) && $util.isInteger(m.entityId.high)))
                    return "entityId: integer|Long expected";
            }
            if (m.x != null && m.hasOwnProperty("x")) {
                if (!$util.isInteger(m.x))
                    return "x: integer expected";
            }
            if (m.y != null && m.hasOwnProperty("y")) {
                if (!$util.isInteger(m.y))
                    return "y: integer expected";
            }
            if (m.w != null && m.hasOwnProperty("w")) {
                if (!$util.isInteger(m.w))
                    return "w: integer expected";
            }
            if (m.h != null && m.hasOwnProperty("h")) {
                if (!$util.isInteger(m.h))
                    return "h: integer expected";
            }
            if (m.primaryColor != null && m.hasOwnProperty("primaryColor")) {
                if (!$util.isInteger(m.primaryColor))
                    return "primaryColor: integer expected";
            }
            if (m.definition != null && m.hasOwnProperty("definition")) {
                if (!$util.isString(m.definition))
                    return "definition: string expected";
            }
            if (m.initialState != null && m.hasOwnProperty("initialState")) {
                if (!$util.isString(m.initialState))
                    return "initialState: string expected";
            }
            return null;
        };

        EntitySpawn.fromObject = function fromObject(d) {
            if (d instanceof $root.greyvarproto.EntitySpawn)
                return d;
            var m = new $root.greyvarproto.EntitySpawn();
            if (d.entityId != null) {
                if ($util.Long)
                    (m.entityId = $util.Long.fromValue(d.entityId)).unsigned = false;
                else if (typeof d.entityId === "string")
                    m.entityId = parseInt(d.entityId, 10);
                else if (typeof d.entityId === "number")
                    m.entityId = d.entityId;
                else if (typeof d.entityId === "object")
                    m.entityId = new $util.LongBits(d.entityId.low >>> 0, d.entityId.high >>> 0).toNumber();
            }
            if (d.x != null) {
                m.x = d.x | 0;
            }
            if (d.y != null) {
                m.y = d.y | 0;
            }
            if (d.w != null) {
                m.w = d.w | 0;
            }
            if (d.h != null) {
                m.h = d.h | 0;
            }
            if (d.primaryColor != null) {
                m.primaryColor = d.primaryColor >>> 0;
            }
            if (d.definition != null) {
                m.definition = String(d.definition);
            }
            if (d.initialState != null) {
                m.initialState = String(d.initialState);
            }
            return m;
        };

        EntitySpawn.toObject = function toObject(m, o) {
            if (!o)
                o = {};
            var d = {};
            if (o.defaults) {
                if ($util.Long) {
                    var n = new $util.Long(0, 0, false);
                    d.entityId = o.longs === String ? n.toString() : o.longs === Number ? n.toNumber() : n;
                } else
                    d.entityId = o.longs === String ? "0" : 0;
                d.x = 0;
                d.y = 0;
                d.w = 0;
                d.h = 0;
                d.primaryColor = 0;
                d.definition = "";
                d.initialState = "";
            }
            if (m.entityId != null && m.hasOwnProperty("entityId")) {
                if (typeof m.entityId === "number")
                    d.entityId = o.longs === String ? String(m.entityId) : m.entityId;
                else
                    d.entityId = o.longs === String ? $util.Long.prototype.toString.call(m.entityId) : o.longs === Number ? new $util.LongBits(m.entityId.low >>> 0, m.entityId.high >>> 0).toNumber() : m.entityId;
            }
            if (m.x != null && m.hasOwnProperty("x")) {
                d.x = m.x;
            }
            if (m.y != null && m.hasOwnProperty("y")) {
                d.y = m.y;
            }
            if (m.w != null && m.hasOwnProperty("w")) {
                d.w = m.w;
            }
            if (m.h != null && m.hasOwnProperty("h")) {
                d.h = m.h;
            }
            if (m.primaryColor != null && m.hasOwnProperty("primaryColor")) {
                d.primaryColor = m.primaryColor;
            }
            if (m.definition != null && m.hasOwnProperty("definition")) {
                d.definition = m.definition;
            }
            if (m.initialState != null && m.hasOwnProperty("initialState")) {
                d.initialState = m.initialState;
            }
            return d;
        };

        EntitySpawn.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        return EntitySpawn;
    })();

    greyvarproto.EntityDefinition = (function() {

        function EntityDefinition(p) {
            this.states = [];
            if (p)
                for (var ks = Object.keys(p), i = 0; i < ks.length; ++i)
                    if (p[ks[i]] != null)
                        this[ks[i]] = p[ks[i]];
        }

        EntityDefinition.prototype.name = "";
        EntityDefinition.prototype.states = $util.emptyArray;
        EntityDefinition.prototype.texture = "";

        EntityDefinition.create = function create(properties) {
            return new EntityDefinition(properties);
        };

        EntityDefinition.encode = function encode(m, w) {
            if (!w)
                w = $Writer.create();
            if (m.name != null && Object.hasOwnProperty.call(m, "name"))
                w.uint32(10).string(m.name);
            if (m.states != null && m.states.length) {
                for (var i = 0; i < m.states.length; ++i)
                    $root.greyvarproto.EntityState.encode(m.states[i], w.uint32(26).fork()).ldelim();
            }
            if (m.texture != null && Object.hasOwnProperty.call(m, "texture"))
                w.uint32(34).string(m.texture);
            return w;
        };

        EntityDefinition.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        EntityDefinition.decode = function decode(r, l) {
            if (!(r instanceof $Reader))
                r = $Reader.create(r);
            var c = l === undefined ? r.len : r.pos + l, m = new $root.greyvarproto.EntityDefinition();
            while (r.pos < c) {
                var t = r.uint32();
                switch (t >>> 3) {
                case 1:
                    m.name = r.string();
                    break;
                case 3:
                    if (!(m.states && m.states.length))
                        m.states = [];
                    m.states.push($root.greyvarproto.EntityState.decode(r, r.uint32()));
                    break;
                case 4:
                    m.texture = r.string();
                    break;
                default:
                    r.skipType(t & 7);
                    break;
                }
            }
            return m;
        };

        EntityDefinition.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        EntityDefinition.verify = function verify(m) {
            if (typeof m !== "object" || m === null)
                return "object expected";
            if (m.name != null && m.hasOwnProperty("name")) {
                if (!$util.isString(m.name))
                    return "name: string expected";
            }
            if (m.states != null && m.hasOwnProperty("states")) {
                if (!Array.isArray(m.states))
                    return "states: array expected";
                for (var i = 0; i < m.states.length; ++i) {
                    {
                        var e = $root.greyvarproto.EntityState.verify(m.states[i]);
                        if (e)
                            return "states." + e;
                    }
                }
            }
            if (m.texture != null && m.hasOwnProperty("texture")) {
                if (!$util.isString(m.texture))
                    return "texture: string expected";
            }
            return null;
        };

        EntityDefinition.fromObject = function fromObject(d) {
            if (d instanceof $root.greyvarproto.EntityDefinition)
                return d;
            var m = new $root.greyvarproto.EntityDefinition();
            if (d.name != null) {
                m.name = String(d.name);
            }
            if (d.states) {
                if (!Array.isArray(d.states))
                    throw TypeError(".greyvarproto.EntityDefinition.states: array expected");
                m.states = [];
                for (var i = 0; i < d.states.length; ++i) {
                    if (typeof d.states[i] !== "object")
                        throw TypeError(".greyvarproto.EntityDefinition.states: object expected");
                    m.states[i] = $root.greyvarproto.EntityState.fromObject(d.states[i]);
                }
            }
            if (d.texture != null) {
                m.texture = String(d.texture);
            }
            return m;
        };

        EntityDefinition.toObject = function toObject(m, o) {
            if (!o)
                o = {};
            var d = {};
            if (o.arrays || o.defaults) {
                d.states = [];
            }
            if (o.defaults) {
                d.name = "";
                d.texture = "";
            }
            if (m.name != null && m.hasOwnProperty("name")) {
                d.name = m.name;
            }
            if (m.states && m.states.length) {
                d.states = [];
                for (var j = 0; j < m.states.length; ++j) {
                    d.states[j] = $root.greyvarproto.EntityState.toObject(m.states[j], o);
                }
            }
            if (m.texture != null && m.hasOwnProperty("texture")) {
                d.texture = m.texture;
            }
            return d;
        };

        EntityDefinition.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        return EntityDefinition;
    })();

    greyvarproto.EntityState = (function() {

        function EntityState(p) {
            this.frames = [];
            this.holds = [];
            if (p)
                for (var ks = Object.keys(p), i = 0; i < ks.length; ++i)
                    if (p[ks[i]] != null)
                        this[ks[i]] = p[ks[i]];
        }

        EntityState.prototype.name = "";
        EntityState.prototype.frames = $util.emptyArray;
        EntityState.prototype.holds = $util.emptyArray;

        EntityState.create = function create(properties) {
            return new EntityState(properties);
        };

        EntityState.encode = function encode(m, w) {
            if (!w)
                w = $Writer.create();
            if (m.name != null && Object.hasOwnProperty.call(m, "name"))
                w.uint32(10).string(m.name);
            if (m.frames != null && m.frames.length) {
                w.uint32(18).fork();
                for (var i = 0; i < m.frames.length; ++i)
                    w.int32(m.frames[i]);
                w.ldelim();
            }
            if (m.holds != null && m.holds.length) {
                w.uint32(26).fork();
                for (var i = 0; i < m.holds.length; ++i)
                    w.int32(m.holds[i]);
                w.ldelim();
            }
            return w;
        };

        EntityState.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        EntityState.decode = function decode(r, l) {
            if (!(r instanceof $Reader))
                r = $Reader.create(r);
            var c = l === undefined ? r.len : r.pos + l, m = new $root.greyvarproto.EntityState();
            while (r.pos < c) {
                var t = r.uint32();
                switch (t >>> 3) {
                case 1:
                    m.name = r.string();
                    break;
                case 2:
                    if (!(m.frames && m.frames.length))
                        m.frames = [];
                    if ((t & 7) === 2) {
                        var c2 = r.uint32() + r.pos;
                        while (r.pos < c2)
                            m.frames.push(r.int32());
                    } else
                        m.frames.push(r.int32());
                    break;
                case 3:
                    if (!(m.holds && m.holds.length))
                        m.holds = [];
                    if ((t & 7) === 2) {
                        var c3 = r.uint32() + r.pos;
                        while (r.pos < c3)
                            m.holds.push(r.int32());
                    } else
                        m.holds.push(r.int32());
                    break;
                default:
                    r.skipType(t & 7);
                    break;
                }
            }
            return m;
        };

        EntityState.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        EntityState.verify = function verify(m) {
            if (typeof m !== "object" || m === null)
                return "object expected";
            if (m.name != null && m.hasOwnProperty("name")) {
                if (!$util.isString(m.name))
                    return "name: string expected";
            }
            if (m.frames != null && m.hasOwnProperty("frames")) {
                if (!Array.isArray(m.frames))
                    return "frames: array expected";
                for (var i = 0; i < m.frames.length; ++i) {
                    if (!$util.isInteger(m.frames[i]))
                        return "frames: integer[] expected";
                }
            }
            if (m.holds != null && m.hasOwnProperty("holds")) {
                if (!Array.isArray(m.holds))
                    return "holds: array expected";
                for (var i = 0; i < m.holds.length; ++i) {
                    if (!$util.isInteger(m.holds[i]))
                        return "holds: integer[] expected";
                }
            }
            return null;
        };

        EntityState.fromObject = function fromObject(d) {
            if (d instanceof $root.greyvarproto.EntityState)
                return d;
            var m = new $root.greyvarproto.EntityState();
            if (d.name != null) {
                m.name = String(d.name);
            }
            if (d.frames) {
                if (!Array.isArray(d.frames))
                    throw TypeError(".greyvarproto.EntityState.frames: array expected");
                m.frames = [];
                for (var i = 0; i < d.frames.length; ++i) {
                    m.frames[i] = d.frames[i] | 0;
                }
            }
            if (d.holds) {
                if (!Array.isArray(d.holds))
                    throw TypeError(".greyvarproto.EntityState.holds: array expected");
                m.holds = [];
                for (var i = 0; i < d.holds.length; ++i) {
                    m.holds[i] = d.holds[i] | 0;
                }
            }
            return m;
        };

        EntityState.toObject = function toObject(m, o) {
            if (!o)
                o = {};
            var d = {};
            if (o.arrays || o.defaults) {
                d.frames = [];
                d.holds = [];
            }
            if (o.defaults) {
                d.name = "";
            }
            if (m.name != null && m.hasOwnProperty("name")) {
                d.name = m.name;
            }
            if (m.frames && m.frames.length) {
                d.frames = [];
                for (var j = 0; j < m.frames.length; ++j) {
                    d.frames[j] = m.frames[j];
                }
            }
            if (m.holds && m.holds.length) {
                d.holds = [];
                for (var j = 0; j < m.holds.length; ++j) {
                    d.holds[j] = m.holds[j];
                }
            }
            return d;
        };

        EntityState.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        return EntityState;
    })();

    greyvarproto.PlayerQuit = (function() {

        function PlayerQuit(p) {
            if (p)
                for (var ks = Object.keys(p), i = 0; i < ks.length; ++i)
                    if (p[ks[i]] != null)
                        this[ks[i]] = p[ks[i]];
        }

        PlayerQuit.prototype.playerId = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        PlayerQuit.create = function create(properties) {
            return new PlayerQuit(properties);
        };

        PlayerQuit.encode = function encode(m, w) {
            if (!w)
                w = $Writer.create();
            if (m.playerId != null && Object.hasOwnProperty.call(m, "playerId"))
                w.uint32(8).int64(m.playerId);
            return w;
        };

        PlayerQuit.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        PlayerQuit.decode = function decode(r, l) {
            if (!(r instanceof $Reader))
                r = $Reader.create(r);
            var c = l === undefined ? r.len : r.pos + l, m = new $root.greyvarproto.PlayerQuit();
            while (r.pos < c) {
                var t = r.uint32();
                switch (t >>> 3) {
                case 1:
                    m.playerId = r.int64();
                    break;
                default:
                    r.skipType(t & 7);
                    break;
                }
            }
            return m;
        };

        PlayerQuit.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        PlayerQuit.verify = function verify(m) {
            if (typeof m !== "object" || m === null)
                return "object expected";
            if (m.playerId != null && m.hasOwnProperty("playerId")) {
                if (!$util.isInteger(m.playerId) && !(m.playerId && $util.isInteger(m.playerId.low) && $util.isInteger(m.playerId.high)))
                    return "playerId: integer|Long expected";
            }
            return null;
        };

        PlayerQuit.fromObject = function fromObject(d) {
            if (d instanceof $root.greyvarproto.PlayerQuit)
                return d;
            var m = new $root.greyvarproto.PlayerQuit();
            if (d.playerId != null) {
                if ($util.Long)
                    (m.playerId = $util.Long.fromValue(d.playerId)).unsigned = false;
                else if (typeof d.playerId === "string")
                    m.playerId = parseInt(d.playerId, 10);
                else if (typeof d.playerId === "number")
                    m.playerId = d.playerId;
                else if (typeof d.playerId === "object")
                    m.playerId = new $util.LongBits(d.playerId.low >>> 0, d.playerId.high >>> 0).toNumber();
            }
            return m;
        };

        PlayerQuit.toObject = function toObject(m, o) {
            if (!o)
                o = {};
            var d = {};
            if (o.defaults) {
                if ($util.Long) {
                    var n = new $util.Long(0, 0, false);
                    d.playerId = o.longs === String ? n.toString() : o.longs === Number ? n.toNumber() : n;
                } else
                    d.playerId = o.longs === String ? "0" : 0;
            }
            if (m.playerId != null && m.hasOwnProperty("playerId")) {
                if (typeof m.playerId === "number")
                    d.playerId = o.longs === String ? String(m.playerId) : m.playerId;
                else
                    d.playerId = o.longs === String ? $util.Long.prototype.toString.call(m.playerId) : o.longs === Number ? new $util.LongBits(m.playerId.low >>> 0, m.playerId.high >>> 0).toNumber() : m.playerId;
            }
            return d;
        };

        PlayerQuit.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        return PlayerQuit;
    })();

    greyvarproto.PlayerAlreadyHere = (function() {

        function PlayerAlreadyHere(p) {
            if (p)
                for (var ks = Object.keys(p), i = 0; i < ks.length; ++i)
                    if (p[ks[i]] != null)
                        this[ks[i]] = p[ks[i]];
        }

        PlayerAlreadyHere.prototype.playerId = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        PlayerAlreadyHere.create = function create(properties) {
            return new PlayerAlreadyHere(properties);
        };

        PlayerAlreadyHere.encode = function encode(m, w) {
            if (!w)
                w = $Writer.create();
            if (m.playerId != null && Object.hasOwnProperty.call(m, "playerId"))
                w.uint32(8).int64(m.playerId);
            return w;
        };

        PlayerAlreadyHere.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        PlayerAlreadyHere.decode = function decode(r, l) {
            if (!(r instanceof $Reader))
                r = $Reader.create(r);
            var c = l === undefined ? r.len : r.pos + l, m = new $root.greyvarproto.PlayerAlreadyHere();
            while (r.pos < c) {
                var t = r.uint32();
                switch (t >>> 3) {
                case 1:
                    m.playerId = r.int64();
                    break;
                default:
                    r.skipType(t & 7);
                    break;
                }
            }
            return m;
        };

        PlayerAlreadyHere.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        PlayerAlreadyHere.verify = function verify(m) {
            if (typeof m !== "object" || m === null)
                return "object expected";
            if (m.playerId != null && m.hasOwnProperty("playerId")) {
                if (!$util.isInteger(m.playerId) && !(m.playerId && $util.isInteger(m.playerId.low) && $util.isInteger(m.playerId.high)))
                    return "playerId: integer|Long expected";
            }
            return null;
        };

        PlayerAlreadyHere.fromObject = function fromObject(d) {
            if (d instanceof $root.greyvarproto.PlayerAlreadyHere)
                return d;
            var m = new $root.greyvarproto.PlayerAlreadyHere();
            if (d.playerId != null) {
                if ($util.Long)
                    (m.playerId = $util.Long.fromValue(d.playerId)).unsigned = false;
                else if (typeof d.playerId === "string")
                    m.playerId = parseInt(d.playerId, 10);
                else if (typeof d.playerId === "number")
                    m.playerId = d.playerId;
                else if (typeof d.playerId === "object")
                    m.playerId = new $util.LongBits(d.playerId.low >>> 0, d.playerId.high >>> 0).toNumber();
            }
            return m;
        };

        PlayerAlreadyHere.toObject = function toObject(m, o) {
            if (!o)
                o = {};
            var d = {};
            if (o.defaults) {
                if ($util.Long) {
                    var n = new $util.Long(0, 0, false);
                    d.playerId = o.longs === String ? n.toString() : o.longs === Number ? n.toNumber() : n;
                } else
                    d.playerId = o.longs === String ? "0" : 0;
            }
            if (m.playerId != null && m.hasOwnProperty("playerId")) {
                if (typeof m.playerId === "number")
                    d.playerId = o.longs === String ? String(m.playerId) : m.playerId;
                else
                    d.playerId = o.longs === String ? $util.Long.prototype.toString.call(m.playerId) : o.longs === Number ? new $util.LongBits(m.playerId.low >>> 0, m.playerId.high >>> 0).toNumber() : m.playerId;
            }
            return d;
        };

        PlayerAlreadyHere.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        return PlayerAlreadyHere;
    })();

    greyvarproto.PlayerJoined = (function() {

        function PlayerJoined(p) {
            if (p)
                for (var ks = Object.keys(p), i = 0; i < ks.length; ++i)
                    if (p[ks[i]] != null)
                        this[ks[i]] = p[ks[i]];
        }

        PlayerJoined.prototype.username = "";

        PlayerJoined.create = function create(properties) {
            return new PlayerJoined(properties);
        };

        PlayerJoined.encode = function encode(m, w) {
            if (!w)
                w = $Writer.create();
            if (m.username != null && Object.hasOwnProperty.call(m, "username"))
                w.uint32(10).string(m.username);
            return w;
        };

        PlayerJoined.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        PlayerJoined.decode = function decode(r, l) {
            if (!(r instanceof $Reader))
                r = $Reader.create(r);
            var c = l === undefined ? r.len : r.pos + l, m = new $root.greyvarproto.PlayerJoined();
            while (r.pos < c) {
                var t = r.uint32();
                switch (t >>> 3) {
                case 1:
                    m.username = r.string();
                    break;
                default:
                    r.skipType(t & 7);
                    break;
                }
            }
            return m;
        };

        PlayerJoined.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        PlayerJoined.verify = function verify(m) {
            if (typeof m !== "object" || m === null)
                return "object expected";
            if (m.username != null && m.hasOwnProperty("username")) {
                if (!$util.isString(m.username))
                    return "username: string expected";
            }
            return null;
        };

        PlayerJoined.fromObject = function fromObject(d) {
            if (d instanceof $root.greyvarproto.PlayerJoined)
                return d;
            var m = new $root.greyvarproto.PlayerJoined();
            if (d.username != null) {
                m.username = String(d.username);
            }
            return m;
        };

        PlayerJoined.toObject = function toObject(m, o) {
            if (!o)
                o = {};
            var d = {};
            if (o.defaults) {
                d.username = "";
            }
            if (m.username != null && m.hasOwnProperty("username")) {
                d.username = m.username;
            }
            return d;
        };

        PlayerJoined.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        return PlayerJoined;
    })();

    greyvarproto.PlayerYou = (function() {

        function PlayerYou(p) {
            if (p)
                for (var ks = Object.keys(p), i = 0; i < ks.length; ++i)
                    if (p[ks[i]] != null)
                        this[ks[i]] = p[ks[i]];
        }

        PlayerYou.prototype.playerId = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        PlayerYou.create = function create(properties) {
            return new PlayerYou(properties);
        };

        PlayerYou.encode = function encode(m, w) {
            if (!w)
                w = $Writer.create();
            if (m.playerId != null && Object.hasOwnProperty.call(m, "playerId"))
                w.uint32(8).int64(m.playerId);
            return w;
        };

        PlayerYou.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        PlayerYou.decode = function decode(r, l) {
            if (!(r instanceof $Reader))
                r = $Reader.create(r);
            var c = l === undefined ? r.len : r.pos + l, m = new $root.greyvarproto.PlayerYou();
            while (r.pos < c) {
                var t = r.uint32();
                switch (t >>> 3) {
                case 1:
                    m.playerId = r.int64();
                    break;
                default:
                    r.skipType(t & 7);
                    break;
                }
            }
            return m;
        };

        PlayerYou.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        PlayerYou.verify = function verify(m) {
            if (typeof m !== "object" || m === null)
                return "object expected";
            if (m.playerId != null && m.hasOwnProperty("playerId")) {
                if (!$util.isInteger(m.playerId) && !(m.playerId && $util.isInteger(m.playerId.low) && $util.isInteger(m.playerId.high)))
                    return "playerId: integer|Long expected";
            }
            return null;
        };

        PlayerYou.fromObject = function fromObject(d) {
            if (d instanceof $root.greyvarproto.PlayerYou)
                return d;
            var m = new $root.greyvarproto.PlayerYou();
            if (d.playerId != null) {
                if ($util.Long)
                    (m.playerId = $util.Long.fromValue(d.playerId)).unsigned = false;
                else if (typeof d.playerId === "string")
                    m.playerId = parseInt(d.playerId, 10);
                else if (typeof d.playerId === "number")
                    m.playerId = d.playerId;
                else if (typeof d.playerId === "object")
                    m.playerId = new $util.LongBits(d.playerId.low >>> 0, d.playerId.high >>> 0).toNumber();
            }
            return m;
        };

        PlayerYou.toObject = function toObject(m, o) {
            if (!o)
                o = {};
            var d = {};
            if (o.defaults) {
                if ($util.Long) {
                    var n = new $util.Long(0, 0, false);
                    d.playerId = o.longs === String ? n.toString() : o.longs === Number ? n.toNumber() : n;
                } else
                    d.playerId = o.longs === String ? "0" : 0;
            }
            if (m.playerId != null && m.hasOwnProperty("playerId")) {
                if (typeof m.playerId === "number")
                    d.playerId = o.longs === String ? String(m.playerId) : m.playerId;
                else
                    d.playerId = o.longs === String ? $util.Long.prototype.toString.call(m.playerId) : o.longs === Number ? new $util.LongBits(m.playerId.low >>> 0, m.playerId.high >>> 0).toNumber() : m.playerId;
            }
            return d;
        };

        PlayerYou.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        return PlayerYou;
    })();

    greyvarproto.Tile = (function() {

        function Tile(p) {
            if (p)
                for (var ks = Object.keys(p), i = 0; i < ks.length; ++i)
                    if (p[ks[i]] != null)
                        this[ks[i]] = p[ks[i]];
        }

        Tile.prototype.col = 0;
        Tile.prototype.row = 0;
        Tile.prototype.tex = "";
        Tile.prototype.flipH = false;
        Tile.prototype.flipV = false;
        Tile.prototype.rot = 0;

        Tile.create = function create(properties) {
            return new Tile(properties);
        };

        Tile.encode = function encode(m, w) {
            if (!w)
                w = $Writer.create();
            if (m.col != null && Object.hasOwnProperty.call(m, "col"))
                w.uint32(8).uint32(m.col);
            if (m.row != null && Object.hasOwnProperty.call(m, "row"))
                w.uint32(16).uint32(m.row);
            if (m.tex != null && Object.hasOwnProperty.call(m, "tex"))
                w.uint32(26).string(m.tex);
            if (m.flipH != null && Object.hasOwnProperty.call(m, "flipH"))
                w.uint32(32).bool(m.flipH);
            if (m.flipV != null && Object.hasOwnProperty.call(m, "flipV"))
                w.uint32(40).bool(m.flipV);
            if (m.rot != null && Object.hasOwnProperty.call(m, "rot"))
                w.uint32(48).int32(m.rot);
            return w;
        };

        Tile.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        Tile.decode = function decode(r, l) {
            if (!(r instanceof $Reader))
                r = $Reader.create(r);
            var c = l === undefined ? r.len : r.pos + l, m = new $root.greyvarproto.Tile();
            while (r.pos < c) {
                var t = r.uint32();
                switch (t >>> 3) {
                case 1:
                    m.col = r.uint32();
                    break;
                case 2:
                    m.row = r.uint32();
                    break;
                case 3:
                    m.tex = r.string();
                    break;
                case 4:
                    m.flipH = r.bool();
                    break;
                case 5:
                    m.flipV = r.bool();
                    break;
                case 6:
                    m.rot = r.int32();
                    break;
                default:
                    r.skipType(t & 7);
                    break;
                }
            }
            return m;
        };

        Tile.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        Tile.verify = function verify(m) {
            if (typeof m !== "object" || m === null)
                return "object expected";
            if (m.col != null && m.hasOwnProperty("col")) {
                if (!$util.isInteger(m.col))
                    return "col: integer expected";
            }
            if (m.row != null && m.hasOwnProperty("row")) {
                if (!$util.isInteger(m.row))
                    return "row: integer expected";
            }
            if (m.tex != null && m.hasOwnProperty("tex")) {
                if (!$util.isString(m.tex))
                    return "tex: string expected";
            }
            if (m.flipH != null && m.hasOwnProperty("flipH")) {
                if (typeof m.flipH !== "boolean")
                    return "flipH: boolean expected";
            }
            if (m.flipV != null && m.hasOwnProperty("flipV")) {
                if (typeof m.flipV !== "boolean")
                    return "flipV: boolean expected";
            }
            if (m.rot != null && m.hasOwnProperty("rot")) {
                if (!$util.isInteger(m.rot))
                    return "rot: integer expected";
            }
            return null;
        };

        Tile.fromObject = function fromObject(d) {
            if (d instanceof $root.greyvarproto.Tile)
                return d;
            var m = new $root.greyvarproto.Tile();
            if (d.col != null) {
                m.col = d.col >>> 0;
            }
            if (d.row != null) {
                m.row = d.row >>> 0;
            }
            if (d.tex != null) {
                m.tex = String(d.tex);
            }
            if (d.flipH != null) {
                m.flipH = Boolean(d.flipH);
            }
            if (d.flipV != null) {
                m.flipV = Boolean(d.flipV);
            }
            if (d.rot != null) {
                m.rot = d.rot | 0;
            }
            return m;
        };

        Tile.toObject = function toObject(m, o) {
            if (!o)
                o = {};
            var d = {};
            if (o.defaults) {
                d.col = 0;
                d.row = 0;
                d.tex = "";
                d.flipH = false;
                d.flipV = false;
                d.rot = 0;
            }
            if (m.col != null && m.hasOwnProperty("col")) {
                d.col = m.col;
            }
            if (m.row != null && m.hasOwnProperty("row")) {
                d.row = m.row;
            }
            if (m.tex != null && m.hasOwnProperty("tex")) {
                d.tex = m.tex;
            }
            if (m.flipH != null && m.hasOwnProperty("flipH")) {
                d.flipH = m.flipH;
            }
            if (m.flipV != null && m.hasOwnProperty("flipV")) {
                d.flipV = m.flipV;
            }
            if (m.rot != null && m.hasOwnProperty("rot")) {
                d.rot = m.rot;
            }
            return d;
        };

        Tile.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        return Tile;
    })();

    greyvarproto.Grid = (function() {

        function Grid(p) {
            this.tiles = [];
            if (p)
                for (var ks = Object.keys(p), i = 0; i < ks.length; ++i)
                    if (p[ks[i]] != null)
                        this[ks[i]] = p[ks[i]];
        }

        Grid.prototype.title = "";
        Grid.prototype.tiles = $util.emptyArray;
        Grid.prototype.rowCount = 0;
        Grid.prototype.colCount = 0;

        Grid.create = function create(properties) {
            return new Grid(properties);
        };

        Grid.encode = function encode(m, w) {
            if (!w)
                w = $Writer.create();
            if (m.title != null && Object.hasOwnProperty.call(m, "title"))
                w.uint32(10).string(m.title);
            if (m.tiles != null && m.tiles.length) {
                for (var i = 0; i < m.tiles.length; ++i)
                    $root.greyvarproto.Tile.encode(m.tiles[i], w.uint32(18).fork()).ldelim();
            }
            if (m.rowCount != null && Object.hasOwnProperty.call(m, "rowCount"))
                w.uint32(24).uint32(m.rowCount);
            if (m.colCount != null && Object.hasOwnProperty.call(m, "colCount"))
                w.uint32(32).uint32(m.colCount);
            return w;
        };

        Grid.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        Grid.decode = function decode(r, l) {
            if (!(r instanceof $Reader))
                r = $Reader.create(r);
            var c = l === undefined ? r.len : r.pos + l, m = new $root.greyvarproto.Grid();
            while (r.pos < c) {
                var t = r.uint32();
                switch (t >>> 3) {
                case 1:
                    m.title = r.string();
                    break;
                case 2:
                    if (!(m.tiles && m.tiles.length))
                        m.tiles = [];
                    m.tiles.push($root.greyvarproto.Tile.decode(r, r.uint32()));
                    break;
                case 3:
                    m.rowCount = r.uint32();
                    break;
                case 4:
                    m.colCount = r.uint32();
                    break;
                default:
                    r.skipType(t & 7);
                    break;
                }
            }
            return m;
        };

        Grid.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        Grid.verify = function verify(m) {
            if (typeof m !== "object" || m === null)
                return "object expected";
            if (m.title != null && m.hasOwnProperty("title")) {
                if (!$util.isString(m.title))
                    return "title: string expected";
            }
            if (m.tiles != null && m.hasOwnProperty("tiles")) {
                if (!Array.isArray(m.tiles))
                    return "tiles: array expected";
                for (var i = 0; i < m.tiles.length; ++i) {
                    {
                        var e = $root.greyvarproto.Tile.verify(m.tiles[i]);
                        if (e)
                            return "tiles." + e;
                    }
                }
            }
            if (m.rowCount != null && m.hasOwnProperty("rowCount")) {
                if (!$util.isInteger(m.rowCount))
                    return "rowCount: integer expected";
            }
            if (m.colCount != null && m.hasOwnProperty("colCount")) {
                if (!$util.isInteger(m.colCount))
                    return "colCount: integer expected";
            }
            return null;
        };

        Grid.fromObject = function fromObject(d) {
            if (d instanceof $root.greyvarproto.Grid)
                return d;
            var m = new $root.greyvarproto.Grid();
            if (d.title != null) {
                m.title = String(d.title);
            }
            if (d.tiles) {
                if (!Array.isArray(d.tiles))
                    throw TypeError(".greyvarproto.Grid.tiles: array expected");
                m.tiles = [];
                for (var i = 0; i < d.tiles.length; ++i) {
                    if (typeof d.tiles[i] !== "object")
                        throw TypeError(".greyvarproto.Grid.tiles: object expected");
                    m.tiles[i] = $root.greyvarproto.Tile.fromObject(d.tiles[i]);
                }
            }
            if (d.rowCount != null) {
                m.rowCount = d.rowCount >>> 0;
            }
            if (d.colCount != null) {
                m.colCount = d.colCount >>> 0;
            }
            return m;
        };

        Grid.toObject = function toObject(m, o) {
            if (!o)
                o = {};
            var d = {};
            if (o.arrays || o.defaults) {
                d.tiles = [];
            }
            if (o.defaults) {
                d.title = "";
                d.rowCount = 0;
                d.colCount = 0;
            }
            if (m.title != null && m.hasOwnProperty("title")) {
                d.title = m.title;
            }
            if (m.tiles && m.tiles.length) {
                d.tiles = [];
                for (var j = 0; j < m.tiles.length; ++j) {
                    d.tiles[j] = $root.greyvarproto.Tile.toObject(m.tiles[j], o);
                }
            }
            if (m.rowCount != null && m.hasOwnProperty("rowCount")) {
                d.rowCount = m.rowCount;
            }
            if (m.colCount != null && m.hasOwnProperty("colCount")) {
                d.colCount = m.colCount;
            }
            return d;
        };

        Grid.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        return Grid;
    })();

    greyvarproto.ConnectionResponse = (function() {

        function ConnectionResponse(p) {
            this.playerAlreadyHere = [];
            if (p)
                for (var ks = Object.keys(p), i = 0; i < ks.length; ++i)
                    if (p[ks[i]] != null)
                        this[ks[i]] = p[ks[i]];
        }

        ConnectionResponse.prototype.serverVersion = "";
        ConnectionResponse.prototype.playerAlreadyHere = $util.emptyArray;

        ConnectionResponse.create = function create(properties) {
            return new ConnectionResponse(properties);
        };

        ConnectionResponse.encode = function encode(m, w) {
            if (!w)
                w = $Writer.create();
            if (m.serverVersion != null && Object.hasOwnProperty.call(m, "serverVersion"))
                w.uint32(10).string(m.serverVersion);
            if (m.playerAlreadyHere != null && m.playerAlreadyHere.length) {
                for (var i = 0; i < m.playerAlreadyHere.length; ++i)
                    $root.greyvarproto.PlayerAlreadyHere.encode(m.playerAlreadyHere[i], w.uint32(42).fork()).ldelim();
            }
            return w;
        };

        ConnectionResponse.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        ConnectionResponse.decode = function decode(r, l) {
            if (!(r instanceof $Reader))
                r = $Reader.create(r);
            var c = l === undefined ? r.len : r.pos + l, m = new $root.greyvarproto.ConnectionResponse();
            while (r.pos < c) {
                var t = r.uint32();
                switch (t >>> 3) {
                case 1:
                    m.serverVersion = r.string();
                    break;
                case 5:
                    if (!(m.playerAlreadyHere && m.playerAlreadyHere.length))
                        m.playerAlreadyHere = [];
                    m.playerAlreadyHere.push($root.greyvarproto.PlayerAlreadyHere.decode(r, r.uint32()));
                    break;
                default:
                    r.skipType(t & 7);
                    break;
                }
            }
            return m;
        };

        ConnectionResponse.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        ConnectionResponse.verify = function verify(m) {
            if (typeof m !== "object" || m === null)
                return "object expected";
            if (m.serverVersion != null && m.hasOwnProperty("serverVersion")) {
                if (!$util.isString(m.serverVersion))
                    return "serverVersion: string expected";
            }
            if (m.playerAlreadyHere != null && m.hasOwnProperty("playerAlreadyHere")) {
                if (!Array.isArray(m.playerAlreadyHere))
                    return "playerAlreadyHere: array expected";
                for (var i = 0; i < m.playerAlreadyHere.length; ++i) {
                    {
                        var e = $root.greyvarproto.PlayerAlreadyHere.verify(m.playerAlreadyHere[i]);
                        if (e)
                            return "playerAlreadyHere." + e;
                    }
                }
            }
            return null;
        };

        ConnectionResponse.fromObject = function fromObject(d) {
            if (d instanceof $root.greyvarproto.ConnectionResponse)
                return d;
            var m = new $root.greyvarproto.ConnectionResponse();
            if (d.serverVersion != null) {
                m.serverVersion = String(d.serverVersion);
            }
            if (d.playerAlreadyHere) {
                if (!Array.isArray(d.playerAlreadyHere))
                    throw TypeError(".greyvarproto.ConnectionResponse.playerAlreadyHere: array expected");
                m.playerAlreadyHere = [];
                for (var i = 0; i < d.playerAlreadyHere.length; ++i) {
                    if (typeof d.playerAlreadyHere[i] !== "object")
                        throw TypeError(".greyvarproto.ConnectionResponse.playerAlreadyHere: object expected");
                    m.playerAlreadyHere[i] = $root.greyvarproto.PlayerAlreadyHere.fromObject(d.playerAlreadyHere[i]);
                }
            }
            return m;
        };

        ConnectionResponse.toObject = function toObject(m, o) {
            if (!o)
                o = {};
            var d = {};
            if (o.arrays || o.defaults) {
                d.playerAlreadyHere = [];
            }
            if (o.defaults) {
                d.serverVersion = "";
            }
            if (m.serverVersion != null && m.hasOwnProperty("serverVersion")) {
                d.serverVersion = m.serverVersion;
            }
            if (m.playerAlreadyHere && m.playerAlreadyHere.length) {
                d.playerAlreadyHere = [];
                for (var j = 0; j < m.playerAlreadyHere.length; ++j) {
                    d.playerAlreadyHere[j] = $root.greyvarproto.PlayerAlreadyHere.toObject(m.playerAlreadyHere[j], o);
                }
            }
            return d;
        };

        ConnectionResponse.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        return ConnectionResponse;
    })();

    greyvarproto.NoResponse = (function() {

        function NoResponse(p) {
            if (p)
                for (var ks = Object.keys(p), i = 0; i < ks.length; ++i)
                    if (p[ks[i]] != null)
                        this[ks[i]] = p[ks[i]];
        }

        NoResponse.create = function create(properties) {
            return new NoResponse(properties);
        };

        NoResponse.encode = function encode(m, w) {
            if (!w)
                w = $Writer.create();
            return w;
        };

        NoResponse.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        NoResponse.decode = function decode(r, l) {
            if (!(r instanceof $Reader))
                r = $Reader.create(r);
            var c = l === undefined ? r.len : r.pos + l, m = new $root.greyvarproto.NoResponse();
            while (r.pos < c) {
                var t = r.uint32();
                switch (t >>> 3) {
                default:
                    r.skipType(t & 7);
                    break;
                }
            }
            return m;
        };

        NoResponse.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        NoResponse.verify = function verify(m) {
            if (typeof m !== "object" || m === null)
                return "object expected";
            return null;
        };

        NoResponse.fromObject = function fromObject(d) {
            if (d instanceof $root.greyvarproto.NoResponse)
                return d;
            return new $root.greyvarproto.NoResponse();
        };

        NoResponse.toObject = function toObject() {
            return {};
        };

        NoResponse.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        return NoResponse;
    })();

    return greyvarproto;
})();

export { $root as default };
