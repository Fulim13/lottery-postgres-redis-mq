create table if not exists inventory(
    id          serial       primary key,                 -- prize id
    name        varchar(20)  not null,                    -- prize name
    description varchar(100) not null default '',         -- prize description
    picture     varchar(200) not null default '',         -- prize image
    price       int          not null default 0,          -- what it is worth
    count       int          not null default 0           -- units in stock
);

comment on table inventory is 'Prize inventory table; all prizes must be given out during a single event.';

-- The id of the blank prize is set explicitly, because the inside Go code, we hardcodes it as database.EMPTY_GIFT.
-- To make load testing more convenient, I intentionally set the inventory count to a large value. In a real-world scenario, the inventory count can be much smaller.
-- A special ID is assigned to represent “Thanks for playing.” By adjusting its count, we can control the probability of it being selected.
insert into inventory (id,name,picture,price,count) values (1,'Thanks for playing','img/face.png',0,1000);


-- The insert above supplied id=1 by hand, which leaves the sequence sitting at 1.
-- Without nudging it forward the next insert would collide on the primary key.
select setval('inventory_id_seq', (select max(id) from inventory));

insert into inventory (name,picture,price,count) values
('Basketball','img/ball.jpeg',100,1000),
('Cup','img/cup.jpeg',80,1000),
('Laptop','img/laptop.jpeg',6000,200),
('Tablet','img/pad.jpg',4000,300),
('Phone','img/phone.jpeg',5000,400),
('Cooking Pot','img/pot.jpeg',120,1000),
('Tea','img/tea.jpeg',90,1000),
('Drone','img/uav.jpeg',400,100),
('Wine','img/wine.jpeg',160,500);

create table if not exists orders(
    id          int          generated always as identity primary key, -- order id
    gift_id     int          not null,                                 -- prize id
    user_id     int          not null,                                 -- user id (There is no login system, so the user id is hardcoded to 1 in code level)
    count       int          not null default 1,                       -- how many were bought
    create_time timestamptz  not null default current_timestamp        -- when the order was created
);
comment on table orders is 'Orders';

create index if not exists idx_user on orders (user_id);
